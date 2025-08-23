package events

import (
	"context"
	"fmt"
	"strings"

	"github.com/JackBekket/GitHelper/internal/reflexia"
	"github.com/google/go-github/v72/github"
	"github.com/rs/zerolog/log"
)

func (s *EventsService) PushEventAnyHandler(ctx context.Context,
	_, _ string,
	event *github.PushEvent,
) error {
	// Webhook context invalidates early
	// TODO: create context with timer
	ctx = context.Background()

	sender := event.GetSender().GetLogin()
	if !s.WhitelistGatekeep(sender) {
		return fmt.Errorf("%s is not allowed", sender)
	}

	repo := event.GetRepo()
	defaultBranch := repo.GetDefaultBranch()
	if defaultBranch == "" {
		return fmt.Errorf("empty event default branch")
	}

	ref := event.GetRef()

	if !(strings.HasPrefix(ref, "refs/heads") && strings.HasSuffix(ref, defaultBranch)) {
		return nil
	}

	username, token, _, err := s.GithubAPI.GetAuthDataByUser(sender)
	if err != nil {
		return fmt.Errorf("get github auth data: %w", err)
	}

	repoFullName := repo.GetFullName()

	jobQueue, loaded := s.ReflexiaService.GetJobQueue(repoFullName)
	jobQueue <- reflexia.ReflectJob{
		Ctx:      ctx,
		RepoURL:  repo.GetHTMLURL(),
		Username: username,
		Token:    token,
	}
	log.Debug().Msgf("Pushed job for %s to the queue, processing as %s", repo.GetHTMLURL(), username)

	if !loaded {
		log.Debug().Msg("No jobs yet, starting the queue")
		go s.ReflexiaService.ProcessReflectQueue(repoFullName, jobQueue)
	}

	return nil
}
