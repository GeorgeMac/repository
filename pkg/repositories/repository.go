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

func (s Service) Repositories(ctxt context.Context, req models.RepositoriesRequest) (repos []models.Repository, err error) {
	var (
		incoming  = make(chan struct{}, req.Count)
		collected = make(chan models.Repository)
		wg        sync.WaitGroup
	)

	for i := 0; i < req.Count; i++ {
		incoming <- struct{}{}

		wg.Add(1)
		go func() {
			defer wg.Done()

			for range incoming {
				for {
					target, err := s.target.Parse("/repository?failRatio=0.5")
					if err != nil {
						continue
					}

					resp, err := s.cli.Get(target.String())
					if err != nil {
						continue
					}

					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						// keep trying until 200 achieved
						continue
					}

					var repo repo
					if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
						continue
					}

					collected <- repo.Repository

					break
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(collected)
	}()

	seen := map[int]struct{}{}

	for {
		select {
		case <-ctxt.Done():
			close(incoming)

			for range collected {
				// waits until collected is closed
			}

			return repos, ctxt.Err()
		case resp := <-collected:
			if _, ok := seen[resp.ID]; ok && req.Unique {
				// try again as this has already been seen
				incoming <- struct{}{}
				continue
			}

			// track that we have now see this ID
			seen[resp.ID] = struct{}{}

			repos = append(repos, resp)

			if len(repos) == req.Count {
				close(incoming)
				return
			}
		}
	}

	return
}

type repo struct {
	Repository models.Repository `json:"repository"`
}
