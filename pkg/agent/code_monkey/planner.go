package code_monkey

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	agent "github.com/JackBekket/GitHelper/pkg/agent/rag"
	"github.com/tmc/langchaingo/llms"
)

const (
	GraphPlanName  = "plan"
	GraphToolName  = "tool"
	GraphSolveName = "solve"
)

const LLMToolName = "LLM"

const PromptGetPlan = `For the following task, make plans that can solve the problem step by step. For each plan, indicate
which external tool together with tool input to retrieve evidence. You can store the evidence into a
variable #E that can be called by later tools. (Plan, #E1, Plan, #E2, Plan, ...)

Tools can be one of the following:
(1) Google[input]: Worker that searches results from Google. Useful when you need to find short
and succinct answers about a specific topic. The input should be a search query.
(2) ` + LLMToolName + `[input]: A pretrained LLM like yourself. Useful when you need to act with general
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
const PromptSolver = `Solve the following task or problem. To solve the problem, we have made step-by-step Plan and \
retrieved corresponding Evidence to each Plan. Use them with caution since long evidence might \
contain irrelevant information.

%s

Now solve the question or task according to provided Evidence above. Respond with the answer
directly with no extra words.

Task: %s
Response:`

type ReWOOStep struct {
	Plan      string
	StepName  string
	Tool      string
	ToolInput string
}

type ReWOO struct {
	Task       string
	PlanString string
	Steps      []ReWOOStep
	Results    map[string]string
	Result     string
}

var RegexPattern *regexp.Regexp = regexp.MustCompile(`Plan:\s*(.+)\s*(#E\d+)\s*=\s*(\w+)\s*\[([^\]]+)\]`)

func (lc LLMContext) GetPlan(ctx context.Context, s interface{}) (interface{}, error) {
	state := s.(ReWOO)
	task := state.Task

	response, err := lc.LLM.GenerateContent(ctx,
		agent.CreateMessageContentHuman(
			fmt.Sprintf(
				"%s\nList of tools:\n%s\n\n%s",
				PromptGetPlan,
				getToolDesc(*lc.Tools),
				task,
			),
		),
	)
	if err != nil {
		return s, err
	}

	result := response.Choices[0].Content
	matches := RegexPattern.FindAllStringSubmatch(result, -1)

	for _, m := range matches {
		state.Steps = append(state.Steps,
			ReWOOStep{
				// m[0] - full match,
				Plan:      m[1],
				StepName:  m[2],
				Tool:      m[3],
				ToolInput: m[4],
			},
		)

	}

	state.PlanString = result

	return state, nil
}

func (lc LLMContext) ToolExecution(ctx context.Context, s interface{}) (interface{}, error) {
	state := s.(ReWOO)

	step := state.Steps[getCurrentTask(state)]

	for stepName, result := range state.Results {
		step.ToolInput = strings.ReplaceAll(step.ToolInput, stepName, result)
	}
	prompt := step.ToolInput
	options := []llms.CallOption{}
	if step.Tool != LLMToolName {
		prompt = fmt.Sprintf(
			"Use tool %s to process the task.\nTask: %s",
			step.Tool,
			prompt,
		)
		options = append(options, llms.WithTools(*lc.Tools))
	}

	response, err := lc.LLM.GenerateContent(ctx,
		agent.CreateMessageContentHuman(
			prompt,
		),
		options...,
	)
	if err != nil {
		return state, err
	}

	if len(state.Results) == 0 {
		state.Results = map[string]string{}
	}

	// Not sure how to actually execute received tool call
	msg := llms.TextParts(llms.ChatMessageTypeAI, response.Choices[0].Content)
	for _, toolCall := range response.Choices[0].ToolCalls {
		msg.Parts = append(msg.Parts, toolCall)
		fmt.Printf("%+v", toolCall)
	}
	state.Results[step.StepName] = fmt.Sprintf("%v", msg.Parts[0])
	return state, err
}

func (lc LLMContext) Solve(ctx context.Context, s interface{}) (interface{}, error) {
	state := s.(ReWOO)

	plan := ""
	for _, step := range state.Steps {
		for stepName, result := range state.Results {
			step.ToolInput = strings.ReplaceAll(step.ToolInput, stepName, result)
			step.StepName = strings.ReplaceAll(step.StepName, stepName, result)
		}
		plan += fmt.Sprintf("Plan: %s\n%s = %s[%s]\n", step.Plan, step.StepName, step.Tool, step.ToolInput)
	}
	response, err := lc.LLM.GenerateContent(ctx,
		agent.CreateMessageContentHuman(
			fmt.Sprintf(PromptSolver, plan, state.Task),
		),
	)
	if err != nil {
		return state, err
	}

	state.Result = response.Choices[0].Content

	return state, nil
}

func Route(ctx context.Context, state interface{}) string {
	if getCurrentTask(state.(ReWOO)) == -1 {
		return GraphSolveName
	} else {
		return GraphToolName
	}
}

func getCurrentTask(state ReWOO) int {
	if len(state.Results) == len(state.Steps) {
		return -1
	}
	return len(state.Results)
}

func getToolDesc(tools []llms.Tool) string {
	desc := ""
	tools = append(tools, llms.Tool{
		Function: &llms.FunctionDefinition{
			Name: "LLM",
			Description: `A pretrained LLM like yourself. Useful when you need to act with general
world knowledge and common sense. Prioritize it when you are confident in solving the problem
yourself. Input can be any instruction.`,
		},
	})
	for idx, tool := range tools {
		desc += fmt.Sprintf("(%d) %s[input]: %s\n", idx, tool.Function.Name, tool.Function.Description)
	}
	return desc
}
