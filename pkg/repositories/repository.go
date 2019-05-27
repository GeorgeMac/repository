package repositories

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/georgemac/repositories/pkg/models"
)

type Service struct {
	cli      *http.Client
	target   *url.URL
	attempts int
}

func New(repositoryServiceAddress string) (*Service, error) {
	url, err := url.Parse(repositoryServiceAddress)
	if err != nil {
		return nil, err
	}

	return &Service{
		cli:      &http.Client{},
		target:   url,
		attempts: 3,
	}, nil
}

func (s Service) Repositories(_ context.Context, req models.RepositoriesRequest) (repos []models.Repository, err error) {
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
				do := func(in *task) error {
					target, err := s.target.Parse("/repository?failRatio=0.5")
					if err != nil {
						return err
					}

					resp, err := s.cli.Get(target.String())
					if err != nil {
						return err
					}

					defer resp.Body.Close()

					var repo repo
					if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
						return err
					}

					in.Result = repo.Repository

					return nil
				}

				// retry failures
				for i := 0; i < s.attempts; i++ {
					in.Err = nil

					if err := do(&in); err == nil {
						break
					}

					in.Err = err
				}

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

		if _, ok := seen[resp.Result.ID]; ok && req.Unique {
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
