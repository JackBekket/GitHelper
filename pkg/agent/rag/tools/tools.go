package tools

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools/duckduckgo"
)

type ToolData struct {
	Definition llms.FunctionDefinition
	Call       func(context.Context, string) (string, error)
}

type ToolsExectutor struct {
	Tools map[string]ToolData
}

var ErrNoSuchTool = errors.New("no such tool")

func (e ToolsExectutor) Execute(ctx context.Context, call llms.ToolCall) (llms.ToolCallResponse, error) {
	response := llms.ToolCallResponse{
		ToolCallID: call.ID,
		Name:       call.FunctionCall.Name,
	}
	toolData, ok := e.Tools[call.FunctionCall.Name]
	if !ok {
		return response, ErrNoSuchTool
	}

	var err error
	response.Content, err = toolData.Call(ctx, call.FunctionCall.Arguments)
	return response, err
}

func GetTools(aiURL, aiToken, dbConnection string) ([]llms.Tool, *ToolsExectutor, error) {
	semanticSearchTool := SemanticSearchTool{
		AIURL:        aiURL,
		AIToken:      aiToken,
		DBConnection: dbConnection,
		MaxResults:   2,
	}
	ddgSearchTool, err := duckduckgo.New(5, duckduckgo.DefaultUserAgent)
	if err != nil {
		return nil, nil, err
	}

	toolsExecutor := ToolsExectutor{
		Tools: map[string]ToolData{
			SemanticSearchDefinition.Name: {
				Definition: SemanticSearchDefinition,
				Call:       semanticSearchTool.Call,
			},
			DDGSearchDefinition.Name: {
				Definition: DDGSearchDefinition,
				Call: func(ctx context.Context, s string) (string, error) {
					ddgSearchArgs := DDGSearchArgs{}
					if err := json.Unmarshal([]byte(s), &ddgSearchArgs); err != nil {
						return "", err
					}
					return ddgSearchTool.Call(ctx, ddgSearchArgs.Query)
				},
			},
		},
	}
	tools := []llms.Tool{}

	for _, toolData := range toolsExecutor.Tools {
		tools = append(tools, llms.Tool{
			Type:     "function",
			Function: &toolData.Definition,
		})
	}

	return tools, &toolsExecutor, nil
}
