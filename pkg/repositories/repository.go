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
		incoming  = make(chan task, req.Count)
		collected = make(chan task)
		wg        sync.WaitGroup
	)

	for i := 0; i < req.Count; i++ {
		incoming <- task{}

		wg.Add(1)
		go func() {
			defer wg.Done()

			for in := range incoming {
				func(in *task) {
					target, err := s.target.Parse("/repository")
					if err != nil {
						in.Err = err

						return
					}

					resp, err := s.cli.Get(target.String())
					if err != nil {
						in.Err = err

						return
					}

					defer resp.Body.Close()

					var repo repo
					if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
						in.Err = err

						return
					}

					in.Result = repo.Repository
				}(&in)

				collected <- in
			}
		}()
	}

	go func() {
		wg.Wait()
		close(collected)
	}()

	seen := map[int]struct{}{}

	for resp := range collected {
		if resp.Err != nil {
			return nil, resp.Err
		}

		if _, ok := seen[resp.Result.ID]; ok {
			// try again as this has already been seen
			incoming <- task{}
			continue
		}

		// track that we have now see this ID
		seen[resp.Result.ID] = struct{}{}

		repos = append(repos, resp.Result)

		if len(repos) == req.Count {
			close(incoming)
			return
		}
	}

	return
}

type repo struct {
	Repository models.Repository `json:"repository"`
}
