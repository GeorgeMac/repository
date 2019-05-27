package repositories

import (
	"context"
	"net/http/httptest"
	"sort"
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

	repoA = models.Repository{1, "foo", today}
	repoB = models.Repository{2, "bar", yesterday}
	repoC = models.Repository{3, "baz", twoDaysAgo}
)

func TestRepositories(t *testing.T) {

	for _, testCase := range []struct {
		Name string
		// available repos
		Repositories []models.Repository
		// inputs
		Request models.RepositoryRequest
		// expectations
		ExpectedRepositories []models.Repository
		ExpectedError        error
	}{
		{
			Name:                 "fetch one repo",
			Repositories:         []models.Repository{repoA, repoB},
			Request:              models.NewRepositoryRequest(),
			ExpectedRepositories: []models.Repository{repoA},
		},
		{
			Name:                 "fetch two repos",
			Repositories:         []models.Repository{repoB, repoC},
			Request:              models.NewRepositoryRequest(models.WithCount(2)),
			ExpectedRepositories: []models.Repository{repoB, repoC},
		},
		{
			Name:                 "fetch three unique repos",
			Repositories:         []models.Repository{repoA, repoB, repoB, repoB, repoC},
			Request:              models.NewRepositoryRequest(models.WithCount(3), models.Unique),
			ExpectedRepositories: []models.Repository{repoA, repoB, repoC},
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			var (
				testService = &repositoryService{
					repositories: testCase.Repositories,
				}
				testServer               = httptest.NewServer(testService)
				repositoriesService, err = New(testServer.URL)
			)

			defer testServer.Close()

			require.Nil(t, err)

			resp, err := repositoriesService.Repositories(context.TODO(), testCase.Request)
			if testCase.ExpectedError != nil {
				require.Equal(t, testCase.ExpectedError, err)
				return
			}

			// sort resp as this is non-deterministic due to concurrency
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].ID < resp[j].ID
			})

			require.Nil(t, testCase.ExpectedError)

			assert.Equal(t, testCase.ExpectedRepositories, resp)
		})
	}
}
