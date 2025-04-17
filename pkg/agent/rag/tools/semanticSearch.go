package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JackBekket/hellper/lib/embeddings"
	"github.com/tmc/langchaingo/llms"
)

var SemanticSearchDefinition = llms.FunctionDefinition{
	Name:        "semanticSearch",
	Description: "Performs semantic search in the vector store of the saved code blobs. Returns matching file contents",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
			"collection": map[string]any{ //TODO: there should NOT exist arguments which called NAME cause it cause COLLISION with actual function name.    .....more like confusion then collision so there are no error
				"type":        "string",
				"description": "name of collection store in which we perform the search",
			},
		},
	},
}

// TODO: Extract query and store parameters from the arguments
// (logic to extract necessary values for SemanticSearch call)
type SemanticSearchArgs struct {
	Query string `json:"query"`
	//Store string `json:"store"`
	//Options []map[string]any `json:"options"`
	Collection string `json:"collection"` //TODO: ALWAYS CHECK THIS JSON REFERENCE WHEN ALTERING VARS
}

type SemanticSearchTool struct {
	AIURL        string
	AIToken      string
	DBConnection string
	MaxResults   int
}

func (s SemanticSearchTool) Call(ctx context.Context, input string) (string, error) {
	semanticSearchArgs := SemanticSearchArgs{}
	response := ""

	if err := json.Unmarshal([]byte(input), &semanticSearchArgs); err != nil {
		return "", err
	}

	store, err := embeddings.GetVectorStoreWithOptions(
		s.AIURL,
		s.AIToken,
		s.DBConnection,
		semanticSearchArgs.Collection,
	)
	if err != nil {
		return response, err
	}

	searchResults, err := embeddings.SemanticSearch(
		semanticSearchArgs.Query,
		s.MaxResults,
		store,
	)

	if err != nil {
		return response, err
	}

	for _, result := range searchResults {
		response += fmt.Sprintf("%s\n", result.PageContent)
	}

	return response, nil
}
