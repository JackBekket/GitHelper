package tools

import (
	"github.com/tmc/langchaingo/llms"
)

var DDGSearchDefinition = llms.FunctionDefinition{
	Name: "search",
	Description: `
	"A wrapper around DuckDuckGo Search."
	"Free search alternative to google and serpapi."
	"Input should be a search query."`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
		},
	},
}

type DDGSearchArgs struct {
	Query string `json:"query"`
}
