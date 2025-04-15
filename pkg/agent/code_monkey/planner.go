package code_monkey

import (
	"context"
	"regexp"

	agent "github.com/JackBekket/GitHelper/pkg/agent/rag"
)

type Plan struct {
    Task string
    Steps []string
    PlanString string
}


/*
type ReWOO struct {
    Task string
    PlanString string
    Steps []string
    Results map[string]string
    Result string
}
*/

var agent_prompt = `For the following task, make plans that can solve the problem step by step. For each plan, indicate \
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

type ReWOO struct {
    Task string `json:"task"`
    PlanString string `json:"plan_string"`
    Steps []string `json:"steps"`
    Results map[string]interface{} `json:"results"`
    Result string `json:"result"`
}


func get_plan(state ReWOO) (ReWOO, error) {
    task := state.Task
   
    rp := regexp.MustCompile(`Plan:\s*(.+)\s*(#E\d+)\s*=\s*(\w+)\s*\[([^\]]+)\]`)

    request := agent_prompt + task
    req :=agent.CreateMessageContentHuman(request)
    //agent.CreateMessageContentSystem(a)
    ctx := context.Background()
    response, err := Model.GenerateContent(ctx, req)
    if err != nil {
        return state, err
    }
    result := response.Choices[0].Content

    //result := promptTemplate.Invoke({"task": task})

    // Find all matches in the sample text
    matches := rp.FindStringSubmatch(result)
    return ReWOO{Steps: matches, PlanString: result}, nil
}