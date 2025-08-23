package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/JackBekket/GitHelper/internal/events"
	githubapi "github.com/JackBekket/GitHelper/internal/github_api"
	"github.com/JackBekket/GitHelper/internal/llm"
	"github.com/JackBekket/GitHelper/internal/reflexia"
	"github.com/Swarmind/libagent/pkg/agent/generic"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/Swarmind/libagent/pkg/tools"
	"github.com/Swarmind/libswarmind/pkg/api"
	"github.com/cbrgm/githubevents/v2/githubevents"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// NewConfig from libagent, loads env from godotenv
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("new config")
	}

	// Github API
	appIdStr := getEnv("APP_ID")
	appId, err := strconv.ParseInt(appIdStr, 10, 64)
	if err != nil {
		log.Fatal().Err(err).Msg("parsing APP_ID env variable")
	}
	pkPath := getEnv("PRIVKEY_PATH")
	if _, err := os.Stat(pkPath); err != nil {
		log.Fatal().Err(err).Msg("could not read file for PRIVKEY_PATH env variable")
	}
	githubAPI := githubapi.GithubAPI{
		AppId:  appId,
		PkPath: pkPath,
	}

	// Libagent/libSwarmind
	agent := generic.Agent{}
	openaiLLM, err := openai.New(
		openai.WithBaseURL(cfg.AIURL),
		openai.WithToken(cfg.AIToken),
		openai.WithModel(cfg.Model),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("new openai api llm")
	}
	agent.LLM = openaiLLM
	toolsExecutor, err := tools.NewToolsExecutor(ctx, cfg, tools.WithToolsWhitelist(
		tools.ReWOOToolDefinition.Name,
		tools.SemanticSearchDefinition.Name,
		tools.DDGSearchDefinition.Name,
		tools.WebReaderDefinition.Name,
	))
	if err != nil {
		log.Fatal().Err(err).Msg("new tools executor")
	}
	agent.ToolsExecutor = toolsExecutor
	llmService := llm.LLMService{
		Agent: agent,
		SwarmindAPI: api.SwarmindAPI{
			Namespace: "GitHellper",
			URL:       os.Getenv("LIBSWARMIND_API_URL"),
			Token:     cfg.AIToken,
			Client:    &http.Client{},
		},
	}

	// Reflexia
	reflexiaService := reflexia.ReflexiaService{
		ReflexiaURL: getEnv("REFLEXIA_API_URL"),
		AIURL:       cfg.AIURL,
		AIToken:     cfg.AIToken,
		Model:       cfg.Model,

		Client: &http.Client{},
	}

	//Events
	eventsService := events.EventsService{
		Whitelist:       strings.Split(os.Getenv("WHITELIST"), ","),
		GithubAPI:       githubAPI,
		LLMService:      llmService,
		ReflexiaService: &reflexiaService,
	}

	// Event handler
	handle := githubevents.New(os.Getenv("WEBHOOK_SECRET_KEY"))
	handle.OnInstallationRepositoriesEventAny(
		eventsService.InstallationRepositoriesEventAnyHandler,
	)
	handle.OnInstallationEventAny(
		eventsService.InstallationEventAnyHandler,
	)
	handle.OnIssuesEventAny(
		eventsService.IssuesEventAnyHandler,
	)
	handle.OnIssueCommentEventAny(
		eventsService.IssueCommentAnyHandler,
	)
	handle.OnPushEventAny(
		eventsService.PushEventAnyHandler,
	)

	http.HandleFunc(getEnv("WEBHOOK_ROUTE"), func(w http.ResponseWriter, r *http.Request) {
		if err := handle.HandleEventRequest(r); err != nil {
			log.Warn().Err(err).Msg("handle github event request")
		}
	})

	if err := http.ListenAndServe(getEnv("LISTEN_ADDR"), nil); err != nil {
		log.Fatal().Err(err).Msg("listen and serve")
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Msgf("environment key %s is empty", key)
	}
	return val
}
