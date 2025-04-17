package code_monkey

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/JackBekket/GitHelper/pkg/agent/rag/tools"
	graph "github.com/JackBekket/langgraphgo/graph/stategraph"
)

type LLMContext struct {
	LLM           *openai.LLM
	Tools         *[]llms.Tool
	ToolsExecutor *tools.ToolsExectutor
}

// This is the main function for this package
func OneShotRun(prompt string, llm openai.LLM, historyState ...llms.MessageContent) string {
	// please add context passing to the top level
	ctx := context.Background()

	/*
		initialState := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, "Below a current conversation between user and helpful AI assistant. You (assistant) should help user in any task he/she ask you to do."),
		}

		if len(historyState) > 0 { // if there are previouse message state then we first load it into message state
			// Access the first element of the slice
			// ... use the history variable as needed
			initialState = append(initialState, historyState...) // load history as initial state
			initialState = append(
				initialState,
				agentState..., // append agent system prompt
			)
			initialState = append(
				initialState,
				llms.TextParts(llms.ChatMessageTypeHuman, prompt), //append user input (!)
			)
		} else {
			initialState = append(
				initialState,
				agentState...,
			)
			initialState = append(initialState,
				llms.TextParts(llms.ChatMessageTypeHuman, prompt),
			)
		}
	*/

	// MAIN WORKFLOW
	tools, _, err := tools.GetTools(
		os.Getenv("AI_URL"),
		os.Getenv("API_TOKEN"),
		os.Getenv("DB_LINK"),
	)
	if err != nil {
		log.Printf("error: %v", err)
		return fmt.Sprintf("error :%v", err)
	}

	workflow := graph.NewStateGraph()
	lc := LLMContext{
		LLM:   &llm,
		Tools: &tools,
	}

	workflow.AddNode("plan", lc.GetPlan)
	/*
		workflow.AddNode("tool", tool_execution)
		workflow.AddNode("solve", solve)
		workflow.AddEdge("plan", "tool")
		workflow.AddEdge("solve", END)
		workflow.AddConditionalEdge("tool", _route)
	*/
	workflow.SetEntryPoint("plan")

	app, err := workflow.Compile()
	if err != nil {
		log.Printf("error: %v", err)
		return fmt.Sprintf("error :%v", err)
	}

	app.Invoke(ctx, nil)

	/*
		response, err := app.Invoke(context.Background(), initialState)
		if err != nil {
			log.Printf("error: %v", err)
			return fmt.Sprintf("error :%v", err)
		}

		lastMsg := response[len(response)-1]
		result := lastMsg.Parts[0]
		resultStr := fmt.Sprintf("%v", result)
		return resultStr
	*/

	return ""
}
