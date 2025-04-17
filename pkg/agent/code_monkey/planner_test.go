package code_monkey_test

import (
	"os"
	"testing"

	"github.com/JackBekket/GitHelper/pkg/agent/rag/tools"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"

	codeMonkey "github.com/JackBekket/GitHelper/pkg/agent/code_monkey"
)

func TestPlanner(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Fatal(err)
	}

	llm, err := openai.New(
		openai.WithToken(os.Getenv("API_TOKEN")),
		openai.WithModel(os.Getenv("MODEL")),
		openai.WithBaseURL(os.Getenv("AI_URL")),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		t.Fatal(err)
	}
	tools, toolsExecutor, err := tools.GetTools(
		os.Getenv("AI_URL"),
		os.Getenv("API_TOKEN"),
		os.Getenv("DB_LINK"),
	)
	if err != nil {
		t.Fatal(err)
	}

	lc := codeMonkey.LLMContext{
		LLM:           llm,
		Tools:         &tools,
		ToolsExecutor: toolsExecutor,
	}

	s, err := lc.GetPlan(t.Context(), codeMonkey.ReWOO{
		Task: `Using semantic search tool, which can search across various code from the project collections find out the telegram library name in the code file contents for the project called "Hellper". Extract it from the given code and use a web search to find the pkg.go.dev documentation for it. Give me the URL for it.`,
	})
	if err != nil {
		t.Fatal(err)
	}
	state := s.(codeMonkey.ReWOO)

	t.Log("Returned steps:")
	for _, step := range state.Steps {
		t.Logf("Step: %+v", step)
	}
	if len(state.Steps) == 0 {
		t.Fatal("empty steps")
	}

	t.Logf("Returned plan string:\n%s", state.PlanString)
	if state.PlanString == "" {
		t.Fatal("empty plan string")
	}

	pLen := 0
	for {
		route := codeMonkey.Route(t.Context(), state)
		if route == codeMonkey.GraphSolveName {
			s, err := lc.Solve(t.Context(), state)
			if err != nil {
				t.Fatal(err)
			}
			state = s.(codeMonkey.ReWOO)
			if state.Result == "" {
				t.Fatal("empty solver result")
			}
			t.Logf("Answer: %s", state.Result)

			break
		} else {
			s, err := lc.ToolExecution(t.Context(), state)
			if err != nil {
				t.Fatal(err)
			}
			state = s.(codeMonkey.ReWOO)
			if len(state.Results) == pLen {
				t.Fatal("no new results")
			}
			pLen = len(state.Results)

			t.Log("Results update:")
			for stepName, result := range state.Results {
				t.Logf("%s: %+v", stepName, result)
			}
		}
	}
}
