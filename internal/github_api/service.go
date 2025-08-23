package githubapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v72/github"
	"github.com/rs/zerolog/log"
)

type GithubAPI struct {
	AppId  int64
	PkPath string
}

func (s GithubAPI) GetAuthDataByUser(owner string) (string, string, *github.Installation, error) {
	transport := http.DefaultTransport

	appsTransport, err := ghinstallation.NewAppsTransportKeyFromFile(transport, s.AppId, s.PkPath)
	if err != nil {
		return "", "", nil, err
	}

	client := github.NewClient(
		&http.Client{
			Transport: appsTransport,
			Timeout:   time.Second * 30,
		},
	)

	if client == nil {
		return "", "", nil, fmt.Errorf("create git client for app")
	}

	installations, _, err := client.Apps.ListInstallations(context.Background(), &github.ListOptions{})
	if err != nil {
		return "", "", nil, fmt.Errorf("list installations: %+v", err)
	}

	for _, installation := range installations {
		log.Debug().Msgf("installation: %+v",
			*installation,
		)
		log.Debug().Msgf("repository selection: %s",
			installation.GetRepositorySelection(),
		)
	}

	var installID int64
	for _, installation := range installations {
		installID = installation.GetID()

		user := installation.GetAccount()
		username := user.GetLogin()
		targetType := installation.GetTargetType()

		log.Debug().Msgf("installed by %s, target type: %s", username, targetType)

		if username != owner {
			continue
		}

		token, _, err := client.Apps.CreateInstallationToken(
			context.Background(),
			installID,
			&github.InstallationTokenOptions{})
		if err != nil {
			return "", "", nil, fmt.Errorf("creating installation token: %+v", err)
		}

		return username, token.GetToken(), installation, nil
	}

	return "", "", nil, fmt.Errorf("client not found for key: %s", owner)
}

func (s GithubAPI) GetClientByUser(owner string) (*github.Client, *github.Installation, error) {
	_, token, installation, err := s.GetAuthDataByUser(owner)
	if err != nil {
		return nil, nil, err
	}

	apiClient := github.NewClient(nil).WithAuthToken(token)
	if apiClient == nil {
		return nil, nil, fmt.Errorf("empty client")
	}

	return apiClient, installation, nil
}
