package repositories

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"../models"
	"github.com/fsouza/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	today      = time.Now().UTC()
	yesterday  = today.Add(-time.Hour * 24).UTC()
	twoDaysAgo = yesterday.Add(-time.Hour * 24).UTC()
)

func TestRepositories(t *testing.T) {
	var (
		repoA       = models.Repository{1, "foo", today}
		repoB       = models.Repository{2, "bar", yesterday}
		repoC       = models.Repository{3, "baz", twoDaysAgo}
		testService = &repositoryService{
			repositories: []models.Repository{
				repoA,
				repoB,
				repoC,
			},
		}
		testServer               = httptest.NewServer(testService)
		repositoriesService, err = New(testServer.URL)
	)

	defer testServer.Close()

	require.Nil(t, err)

	for _, testCase := range []struct {
		Name                 string
		Request              RepositoryRequest
		ExpectedRepositories []models.Repository
		ExpectedError        error
	}{
		{
			Name:                 "fetch one repo",
			Request:              NewRepositoryRequest(),
			ExpectedRepositories: []models.Repository{repoA},
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			resp, err := repositoriesService.Repositories(context.TODO(), testCase.Request)
			if testCase.ExpectedError != nil {
				require.Equal(t, testCase.ExpectedError, err)
				return
			}

			require.Nil(t, testCase.ExpectedError)

			assert.Equal(t, testCase.ExpectedRepositories, resp)
		})
	}
}
