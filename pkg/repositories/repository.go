package repositories

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"../models"
)

type Service struct {
	cli    *http.Client
	target *url.URL
}

func New(repositoryServiceAddress string) (*Service, error) {
	url, err := url.Parse(repositoryServiceAddress)
	if err != nil {
		return nil, err
	}

	return &Service{
		cli:    &http.Client{},
		target: url,
	}, nil
}

func (s Service) Repositories(_ context.Context, req RepositoryRequest) (repos []models.Repository, err error) {
	for i := 0; i < req.Count; i++ {
		target, err := s.target.Parse("/repository")
		if err != nil {
			return nil, err
		}

		resp, err := s.cli.Get(target.String())
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		var repo repo
		if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
			return nil, err
		}

		repos = append(repos, repo.Repository)
	}

	return
}

type repo struct {
	Repository models.Repository `json:"repository"`
}
