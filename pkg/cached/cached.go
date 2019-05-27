package cached

import (
	"context"
	"sync"

	"github.com/georgemac/repositories/pkg/models"
	"github.com/georgemac/repositories/pkg/server"
)

type Service struct {
	proxy    server.RepositoriesService
	cache    []models.Repository
	lookup   map[int]struct{}
	incoming chan []models.Repository
	mu       sync.RWMutex
}

func New(proxy server.RepositoriesService) *Service {
	s := &Service{
		proxy:  proxy,
		lookup: map[int]struct{}{},
		// low tolerance for busy cache
		incoming: make(chan []models.Repository, 5),
	}

	go s.feedCache()

	return s
}

func (s *Service) feedCache() {
	for found := range s.incoming {
		func() {
			s.mu.Lock()
			defer s.mu.Unlock()

			for _, repo := range found {
				// ensure we only remember things we haven't seen before
				if _, ok := s.lookup[repo.ID]; !ok {
					s.cache = append(s.cache, repo)

					s.lookup[repo.ID] = struct{}{}
				}
			}
		}()
	}
}

func (s *Service) Repositories(ctxt context.Context, req models.RepositoriesRequest) ([]models.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res, err := s.proxy.Repositories(ctxt, req)
	if err != nil {
		if err != context.DeadlineExceeded {
			return res, err
		}

		// attempt to write repos to cache
		select {
		case s.incoming <- res:
		default:
			// throw it away if cache writing busy
		}

		candidates := s.cache
		if req.Unique {
			candidates = s.uniqueCandidates(res)
		}

		// given we still have room in the result and we have cached candidates
		for i := 0; len(res) < req.Count && len(candidates) > 0; i++ {
			// can only use candidates once if request is unique
			if req.Unique && i >= len(candidates) {
				break
			}

			// wrap around the slice of cache keys
			res = append(res, candidates[i%len(candidates)])
		}

		return res, nil
	}

	// attempt to write repos to cache
	select {
	case s.incoming <- res:
	default:
		// throw it away if cache writing busy
	}

	return res, nil
}

func (s *Service) uniqueCandidates(found []models.Repository) (unique []models.Repository) {
	foundMap := map[int]struct{}{}
	for _, repo := range found {
		foundMap[repo.ID] = struct{}{}
	}

	for _, candidate := range s.cache {
		if _, ok := foundMap[candidate.ID]; !ok {
			unique = append(unique, candidate)
		}
	}

	return
}
