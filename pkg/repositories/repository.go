package repositories

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

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
	type task struct {
		Result models.Repository
		Err    error
	}

	var (
		collected = make(chan task)
		wg        sync.WaitGroup
	)

	for i := 0; i < req.Count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			target, err := s.target.Parse("/repository")
			if err != nil {
				collected <- task{Err: err}
				return
			}

			resp, err := s.cli.Get(target.String())
			if err != nil {
				collected <- task{Err: err}
				return
			}

			defer resp.Body.Close()

			var repo repo
			if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
				collected <- task{Err: err}
				return
			}

			collected <- task{Result: repo.Repository}
		}()
	}

	go func() {
		wg.Wait()
		close(collected)
	}()

	for task := range collected {
		if task.Err != nil {
			return nil, task.Err
		}

		repos = append(repos, task.Result)
	}

	return
}

type repo struct {
	Repository models.Repository `json:"repository"`
}
