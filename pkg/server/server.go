package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/georgemac/repositories/pkg/models"
)

type RepositoriesService interface {
	Repositories(context.Context, models.RepositoriesRequest) ([]models.Repository, error)
}

type Server struct {
	RepositoriesService RepositoriesService
}

func New(s RepositoriesService) *Server {
	return &Server{s}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := models.NewRepositoriesRequest()

	if v := r.URL.Query().Get("count"); v != "" {
		count, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		models.WithCount(int(count))(&req)
	}

	if v := r.URL.Query().Get("unique"); v == "true" {
		models.Unique(&req)
	}

	ctxt := r.Context()

	if v := r.URL.Query().Get("timeout"); v != "" {
		dur, err := time.ParseDuration(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// configure timeout on repositories call
		var cancel context.CancelFunc
		ctxt, cancel = context.WithTimeout(ctxt, dur)
		defer cancel()
	}

	resp, err := s.RepositoriesService.Repositories(ctxt, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
