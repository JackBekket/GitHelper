package code_monekey

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/JackBekket/GitHelper/pkg/agent/rag/tools"
	graph "github.com/JackBekket/langgraphgo/graph/stategraph"
)

// global var
var Model openai.LLM
var Tools []llms.Tool


// This is the main function for this package
func OneShotRun(prompt string, model openai.LLM, historyState ...llms.MessageContent) string {

	agent_prompt := `For the following task, make plans that can solve the problem step by step. For each plan, indicate \
which external tool together with tool input to retrieve evidence. You can store the evidence into a \
variable #E that can be called by later tools. (Plan, #E1, Plan, #E2, Plan, ...)

Tools can be one of the following:
(1) Google[input]: Worker that searches results from Google. Useful when you need to find short
and succinct answers about a specific topic. The input should be a search query.
(2) LLM[input]: A pretrained LLM like yourself. Useful when you need to act with general
world knowledge and common sense. Prioritize it when you are confident in solving the problem
yourself. Input can be any instruction.

For example,
Task: Thomas, Toby, and Rebecca worked a total of 157 hours in one week. Thomas worked x
hours. Toby worked 10 hours less than twice what Thomas worked, and Rebecca worked 8 hours
less than Toby. How many hours did Rebecca work?
Plan: Given Thomas worked x hours, translate the problem into algebraic expressions and solve
with Wolfram Alpha. #E1 = WolframAlpha[Solve x + (2x − 10) + ((2x − 10) − 8) = 157]
Plan: Find out the number of hours Thomas worked. #E2 = LLM[What is x, given #E1]
Plan: Calculate the number of hours Rebecca worked. #E3 = Calculator[(2 ∗ #E2 − 10) − 8]

Begin! 
Describe your plans with rich details. Each Plan should be followed by only one #E.

Task: `



	agentState := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, agent_prompt),
	}
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

	Tools, _ = tools.GetTools()

	//Tools = tools
	Model = model

	// MAIN WORKFLOW
	workflow := graph.NewStateGraph()

	workflow.AddNode("plan", getPlan)
	workflow.AddNode("tool", tool_execution)
	workflow.AddNode("solve", solve)
	workflow.AddEdge("plan", "tool")
	workflow.AddEdge("solve", END)
	workflow.AddConditionalEdge("tool", _route)
	workflow.SetEntryPoint("plan")

	app, err := workflow.Compile()
	if err != nil {
		log.Printf("error: %v", err)
		return fmt.Sprintf("error :%v", err)
	}

	response, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Printf("error: %v", err)
		return fmt.Sprintf("error :%v", err)
	}

	lastMsg := response[len(response)-1]
	result := lastMsg.Parts[0]
	resultStr := fmt.Sprintf("%v", result)
	return resultStr
}

// AGENT NODE
/** We are telling agent, that it should response withTools, giving it function signatures defined earlier.
  if agent get response from conditional edge like 'yes, use x function with this signatures and this json object as input parameters -- it will match with predefined pointer to semanticSearch function and it will make a toolCall
  then it will append toolCall to message state.
  Agent will recive current stake, make consideration whether or not to use tool and make a call for it
  `shouldSearchDocuments` func will handle this tool call -- it will call semanticSearch function
  Then result of the search tool will go back to agent as a toolResonse in the messages state
*/
func agent(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {

	agentState := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are helpful agent that has access to a semanticSearch tool. Use this tool if user ask to retrive some information from database/collection to provide user with information he/she looking for."),
	}

	/*
		for _,cn := range collection_name {
			collectionState := []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeSystem, "Collection Name: " +cn),
			}
			state := append(agentState,collectionState...)
			agentState = state
		}
	*/

	model := Model // global... should be .env or getting from user context I guess.
	tools := Tools

	/*
		consideration_query := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, "You are decision making agent, which can reply ONLY 'true' or 'false'.Your task is to determine whether or not to call semanticSearch function based on human input. If you see a basic question, return false. If user specified that he desires to use that function, return true. You should ONLY return 'true' or 'false'."),
		}
	*/

	lastMsg := state[len(state)-1]
	if lastMsg.Role == "tool" { // If we catch response from tool then it's second iteration and we simply need to give answer to user using this result
		state = append(state, lastMsg)
		response, err := model.GenerateContent(ctx, state)
		if err != nil {
			return state, err
		}
		msg := llms.TextParts(llms.ChatMessageTypeAI, response.Choices[0].Content)
		state = append(state, msg)
		return state, nil

	} else { // If it is not tool response

		if lastMsg.Role == "human" { //                                            any user request

			// this is consideration stack, it should be placed as separate node.
			/*
				consideration_stack := append(consideration_query, lastMsg)
				//consideration_stack := append(consideration_query, state...)  // this is appending current state, but we actually need only last message here.
				check, err := model.GenerateContent(ctx, consideration_stack) // one punch which determine wheter or not call tools. this is hardcode and probably should be separate part of the graph.
				if err != nil {
					return state, err
				}
				check_txt := fmt.Sprintf(check.Choices[0].Content)
				log.Println("check result: ", check_txt)
			*/
			//	if check_txt == "true" { // tool call required by one-shot agent
			state = append(state, agentState...)
			state = append(state, lastMsg)
			response, err := model.GenerateContent(ctx, state, llms.WithTools(tools)) // AI call tool function.. in this step it just put call in messages stack
			if err != nil {
				return state, err
			}
			msg := llms.TextParts(llms.ChatMessageTypeAI, response.Choices[0].Content)

			if len(response.Choices[0].ToolCalls) > 0 {
				for _, toolCall := range response.Choices[0].ToolCalls {
					if toolCall.FunctionCall.Name == "semanticSearch" { // AI catch that there is a function call in messages, so *now* it actually calls the function.
						msg.Parts = append(msg.Parts, toolCall) // Add result to messages stack
					}
				}
				state = append(state, msg)
				return state, nil
			}
			/*
				} else { // proceed without tools
					response, err := model.GenerateContent(ctx, state)
					if err != nil {
						return state, err
					}
					msg := llms.TextParts(llms.ChatMessageTypeAI, response.Choices[0].Content)
					state = append(state, msg)
					return state, nil
				}
			*/
		} // end if human
		return state, nil
	} // end if not tool response
}

// this function is only HANDLES tool calls, so this is a handler, not a deciding mechanism. agent decide whether or not to call tool in agent func and this func is handling tool call here.
func shouldSearchDocuments(ctx context.Context, state []llms.MessageContent) string {
	// this function (I suppose) can be reworked to work with a *set* of a functions, not just one func.
	lastMsg := state[len(state)-1]
	for _, part := range lastMsg.Parts {
		toolCall, ok := part.(llms.ToolCall)

		if ok && toolCall.FunctionCall.Name == "semanticSearch" {
			log.Printf("agent should use SemanticSearch (embeddings similarity search aka DocumentsSearch)")
			return "semanticSearch"
		}
	}
	return graph.END // never reach this point, should be removed?
}

// This function is performing similarity search in our db vectorstore.
func semanticSearch(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {
	semanticSearchTool := tools.SemanticSearchTool{}
	res, err := semanticSearchTool.Execute(ctx, state)
	if err != nil {
		// Handle the error
		return nil, err
	}
	return res, nil
}
