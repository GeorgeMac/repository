package repositories

import (
	"encoding/json"
	"net/http"
	"sync"

	"../models"
)

type repositoryService struct {
	repositories []models.Repository

	idx int

	mu sync.Mutex
}

func (s *repositoryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp := map[string]models.Repository{"repository": s.repositories[s.idx]}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	s.idx++
	if s.idx >= len(s.repositories) {
		s.idx = 0
	}
}
