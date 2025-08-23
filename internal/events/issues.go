package events

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Swarmind/libswarmind/pkg/api"
	"github.com/google/go-github/v72/github"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
)

func (s *EventsService) IssuesEventAnyHandler(ctx context.Context,
	_, _ string,
	event *github.IssuesEvent,
) error {
	// Webhook context invalidates early
	// TODO: create context with timer
	ctx = context.Background()

	sender := event.GetSender().GetLogin()
	if !s.WhitelistGatekeep(sender) {
		return fmt.Errorf("%s is not allowed", sender)
	}

	repoFullName := event.GetRepo().GetFullName()
	if wg, ok := s.ReflexiaService.ProcessingWg.Load(repoFullName); ok {
		log.Debug().Msg("Waiting until reflexia is done")
		wg.(*sync.WaitGroup).Wait()
		log.Debug().Msg("Starting issue processing")
	}

	client, _, err := s.GithubAPI.GetClientByUser(sender)
	if err != nil {
		return fmt.Errorf("get client for %s: %w", sender, err)
	}

	issue := event.GetIssue()
	issueId := string(issue.GetID())
	issueNumber := issue.GetNumber()
	issueTitle := issue.GetTitle()
	issueBody := issue.GetBody()

	chatName := fmt.Sprintf("%s_%d", strings.ReplaceAll(repoFullName, "/", ":"), issueNumber)

	switch event.GetAction() {
	case "opened":
		// TODO: upd libagent to support reasoning WITH observation, inject a prompt for two-step plan generation
		return s.respond(
			ctx, client, *event.GetRepo(), issueNumber,
			fmt.Sprintf("%s\n\n%s", issueTitle, issueBody),
			issueId, chatName,
		)

	case "edited":
		return s.edit(
			ctx, client, *event.GetRepo(), issueNumber,
			fmt.Sprintf("%s\n\n%s", issueTitle, issueBody),
			issueId, chatName,
		)

	case "deleted":
		err := s.LLMService.SwarmindAPI.DropHistory(ctx, chatName)
		if err != nil {
			return fmt.Errorf("drop history (issue deleted): %w", err)
		}
	}

	return nil
}

func (s *EventsService) IssueCommentAnyHandler(ctx context.Context,
	_, _ string,
	event *github.IssueCommentEvent,
) error {
	// Webhook context invalidates early
	// TODO: create context with timer
	ctx = context.Background()

	sender := event.GetSender().GetLogin()
	if !s.WhitelistGatekeep(sender) {
		return fmt.Errorf("%s is not allowed", sender)
	}

	repoFullName := event.GetRepo().GetFullName()
	if wg, ok := s.ReflexiaService.ProcessingWg.Load(repoFullName); ok {
		log.Debug().Msg("Waiting until reflexia is done")
		wg.(*sync.WaitGroup).Wait()
		log.Debug().Msg("Starting issue processing")
	}

	client, installation, err := s.GithubAPI.GetClientByUser(sender)
	if err != nil {
		return fmt.Errorf("get client for %s: %w", sender, err)
	}

	if event.GetComment().GetUser().GetLogin() == installation.GetAppSlug()+"[bot]" {
		return nil
	}

	issue := event.GetIssue()
	issueNumber := issue.GetNumber()

	comment := event.GetComment()
	commentBody := comment.GetBody()
	commentId := string(comment.GetID())

	chatName := fmt.Sprintf("%s_%d", strings.ReplaceAll(repoFullName, "/", ":"), issueNumber)

	switch event.GetAction() {
	case "created":
		return s.respond(
			ctx, client, *event.GetRepo(), issueNumber,
			commentBody, commentId, chatName,
		)

	case "edited":
		return s.edit(
			ctx, client, *event.GetRepo(), issueNumber,
			commentBody, commentId, chatName,
		)

	case "deleted":
		return s.edit(
			ctx, client, *event.GetRepo(), issueNumber,
			"", commentId, chatName,
		)
	}

	return nil
}

func (s *EventsService) respond(
	ctx context.Context,
	client *github.Client,
	repo github.Repository,
	issueNumber int,
	requestText, requestId, chatName string,
) error {
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	messageContent := llms.TextParts(
		llms.ChatMessageTypeHuman,
		requestText,
	)

	err := s.LLMService.SwarmindAPI.UpdateHistory(ctx, chatName, api.Message{
		ID:      requestId,
		Message: &messageContent,
	})
	if err != nil {
		return fmt.Errorf("update history (request message): %w", err)
	}
	response, err := s.LLMService.GenerateResponse(ctx, chatName)
	if err != nil {
		return err
	}
	issueComment, _, err := client.Issues.CreateComment(ctx, repoOwner, repoName, issueNumber, &github.IssueComment{
		Body: &response,
	})
	if err != nil {
		return fmt.Errorf("create issue comment: %w", err)
	}

	messageContent = llms.TextParts(
		llms.ChatMessageTypeAI,
		response,
	)
	err = s.LLMService.SwarmindAPI.UpdateHistory(ctx, chatName, api.Message{
		ID:      string(issueComment.GetID()),
		Message: &messageContent,
	})
	if err != nil {
		return fmt.Errorf("update history (response message): %w", err)
	}
	return nil
}

func (s *EventsService) edit(
	ctx context.Context,
	client *github.Client,
	repo github.Repository,
	issueNumber int,
	requestText, requestId, chatName string,
) error {
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	var messageContent llms.MessageContent
	if requestText != "" {
		messageContent = llms.TextParts(
			llms.ChatMessageTypeHuman,
			requestText,
		)
	}

	response, invalidated, err := s.LLMService.EditMessageResponse(ctx, chatName, api.Message{
		ID:      requestId,
		Message: &messageContent,
	})

	for _, msg := range invalidated {
		msgId, err := strconv.ParseInt(msg.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("convert message id to int64: %w", err)
		}
		_, err = client.Issues.DeleteComment(ctx, repoOwner, repoName, msgId)
		if err != nil {
			return fmt.Errorf("delete issue comment: %w", err)
		}
	}

	issueComment, _, err := client.Issues.CreateComment(ctx, repoOwner, repoName, issueNumber, &github.IssueComment{
		Body: &response,
	})
	if err != nil {
		return fmt.Errorf("create issue comment: %w", err)
	}

	messageContent = llms.TextParts(
		llms.ChatMessageTypeAI,
		response,
	)
	err = s.LLMService.SwarmindAPI.UpdateHistory(ctx, chatName, api.Message{
		ID:      string(issueComment.GetID()),
		Message: &messageContent,
	})
	if err != nil {
		return fmt.Errorf("update history (response message): %w", err)
	}
	return nil
}
