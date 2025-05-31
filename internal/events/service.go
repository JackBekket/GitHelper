package events

import (
	"slices"

	githubapi "github.com/JackBekket/GitHelper/internal/github_api"
	"github.com/JackBekket/GitHelper/internal/llm"
	"github.com/JackBekket/GitHelper/internal/reflexia"
)

type EventsService struct {
	Whitelist       []string
	GithubAPI       githubapi.GithubAPI
	LLMService      llm.LLMService
	ReflexiaService *reflexia.ReflexiaService
}

func (s *EventsService) WhitelistGatekeep(username string) bool {
	if len(s.Whitelist) == 0 {
		return true
	}

	if slices.Contains(s.Whitelist, username) {
		return true
	}

	return false
}
