package events

import (
	"context"

	"github.com/google/go-github/v72/github"
	"github.com/rs/zerolog/log"
)

func (s *EventsService) InstallationRepositoriesEventAnyHandler(_ context.Context,
	_, _ string,
	event *github.InstallationRepositoriesEvent,
) error {
	sender := event.GetSender().GetLogin()
	allowed := s.WhitelistGatekeep(sender)

	for _, repo := range event.RepositoriesAdded {
		repoFullName := repo.GetFullName()
		log.Info().Msgf("App is installed for %s, allowed: %t", repoFullName, allowed)
	}

	for _, repo := range event.RepositoriesRemoved {
		repoFullName := repo.GetFullName()
		log.Info().Msgf("App is removed for %s, allowed: %t", repoFullName, allowed)
	}

	return nil
}

func (s *EventsService) InstallationEventAnyHandler(_ context.Context,
	_, _ string,
	event *github.InstallationEvent,
) error {
	sender := event.GetSender().GetLogin()
	allowed := s.WhitelistGatekeep(sender)

	for _, repo := range event.Repositories {
		repoFullName := repo.GetFullName()
		log.Info().Msgf("App is installed for %s, allowed: %t", repoFullName, allowed)
	}

	return nil
}
