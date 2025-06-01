package reflexia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/rs/zerolog/log"
)

type ReflexiaService struct {
	ReflexiaURL string
	AIURL       string
	AIToken     string
	Model       string

	Client       *http.Client
	JobQueues    sync.Map
	ProcessingWg sync.Map
}

type ReflectRequest struct {
	AIURL   string `json:"ai_url"`
	AIToken string `json:"ai_token"`
	Model   string `json:"model"`

	RepositoryURL    string `json:"repository_url"`
	RepositoryBranch string `json:"repository_branch,omitempty"`
	GithubUsername   string `json:"github_username,omitempty"`
	GithubToken      string `json:"github_token,omitempty"`

	WithConfigFile string `json:"with_config_file,omitempty"`
	ExactPackages  string `json:"exact_packages,omitempty"`

	CreatePR        bool `json:"create_pr,omitempty"`
	LightCheck      bool `json:"light_check,omitempty"`
	WithFileSummary bool `json:"with_file_summary,omitempty"`
	OverwriteReadme bool `json:"overwrite_readme,omitempty"`

	OverwriteCache bool `json:"overwrite_cache,omitempty"`
	UseEmbeddings  bool `json:"use_embeddings,omitempty"`
}

type ReflectJob struct {
	Ctx      context.Context
	RepoURL  string
	Username string
	Token    string
}

func (s *ReflexiaService) GetJobQueue(repoFullName string) (chan ReflectJob, bool) {
	actual, loaded := s.JobQueues.LoadOrStore(repoFullName, make(chan ReflectJob, 10))
	return actual.(chan ReflectJob), loaded
}

func (s *ReflexiaService) ProcessReflectQueue(repoFullName string, jobChan <-chan ReflectJob) {
	log.Debug().Msgf("Looking for jobs for %s", repoFullName)
	actual, _ := s.ProcessingWg.LoadOrStore(repoFullName, &sync.WaitGroup{})
	wg := actual.(*sync.WaitGroup)
	for job := range jobChan {
		wg.Add(1)
		log.Info().Msgf("Starting job for %s", job.RepoURL)
		err := s.Reflect(job.Ctx, job.RepoURL, job.Username, job.Token)
		if err != nil {
			log.Warn().Err(err).Msgf("reflecting for %s", job.RepoURL)
		}
		wg.Done()
	}
}

func (s *ReflexiaService) Reflect(ctx context.Context, repoURL, username, token string) error {
	requestURL, err := url.JoinPath(s.ReflexiaURL, "reflect")
	if err != nil {
		return fmt.Errorf("failed to process request URL: %w", err)
	}
	jsonBytes, err := json.Marshal(ReflectRequest{
		AIURL:   s.AIURL,
		AIToken: s.AIToken,
		Model:   s.Model,

		RepositoryURL:  repoURL,
		GithubUsername: username,
		GithubToken:    token,

		CreatePR:      true,
		UseEmbeddings: true,
	})
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		requestURL,
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%d status code response: %s", resp.StatusCode, string(body))
	}

	return nil
}
