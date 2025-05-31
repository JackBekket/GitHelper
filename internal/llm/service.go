package llm

import (
	"context"
	"fmt"

	"github.com/Swarmind/libagent/pkg/agent/generic"
	"github.com/Swarmind/libswarmind/pkg/api"
	"github.com/Swarmind/libswarmind/pkg/util"
	"github.com/tmc/langchaingo/llms"
)

type LLMService struct {
	Agent       generic.Agent
	SwarmindAPI api.SwarmindAPI
}

func (s LLMService) GenerateResponse(ctx context.Context, chat string) (string, error) {
	history, err := s.SwarmindAPI.GetHistory(ctx, chat)
	if err != nil {
		return "", fmt.Errorf("get history: %w", err)
	}

	response, err := s.Agent.Run(ctx, util.MessagesToMessageContent(history...))
	if err != nil {
		return "", fmt.Errorf("generate response: %w", err)
	}
	responseTextPart := response.Parts[0].(llms.TextContent)

	return responseTextPart.Text, s.Agent.ToolsExecutor.Cleanup()
}

func (s LLMService) EditMessageResponse(ctx context.Context, chat string, message api.Message) (string, []api.Message, error) {
	history, err := s.SwarmindAPI.GetHistory(ctx, chat)
	if err != nil {
		return "", nil, fmt.Errorf("get history: %w", err)
	}

	invalidated := []api.Message{}
	invalidateLLM := false
	for _, msg := range history {
		if message.ID == msg.ID {
			invalidateLLM = true
		}

		if invalidateLLM {
			if msg.Message.Role == llms.ChatMessageTypeAI {
				continue
			}

			invalidated = append(invalidated, api.Message{
				ID:      msg.ID,
				Message: nil,
			})
		}
	}
	diff := append([]api.Message{message}, invalidated...)
	err = s.SwarmindAPI.UpdateHistory(ctx, chat, diff...)
	if err != nil {
		return "", nil, fmt.Errorf("update history (edit/delete/invalidate): %w", err)
	}

	response, err := s.GenerateResponse(ctx, chat)
	return response, invalidated, err
}
