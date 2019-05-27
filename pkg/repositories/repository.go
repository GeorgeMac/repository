package repositories

import (
	"context"
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

func (s Service) Repositories(context.Context, RepositoryRequest) ([]models.Repository, error) {
	return nil, nil
}
