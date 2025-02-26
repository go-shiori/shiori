package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleLiveness(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("returns build info", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleLiveness(deps, c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response struct {
			Message struct {
				Version string `json:"version"`
				Commit  string `json:"commit"`
				Date    string `json:"date"`
			} `json:"message"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Check build info is populated
		require.Equal(t, model.BuildVersion, response.Message.Version)
		require.Equal(t, model.BuildCommit, response.Message.Commit)
		require.Equal(t, model.BuildDate, response.Message.Date)
	})

	t.Run("handles without auth", func(t *testing.T) {
		// Test that liveness check works without authentication
		c, w := testutil.NewTestWebContext()
		HandleLiveness(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns valid JSON", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleLiveness(deps, c)

		var response struct {
			Message struct {
				Version string `json:"version"`
				Commit  string `json:"commit"`
				Date    string `json:"date"`
			} `json:"message"`
		}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		require.Equal(t, model.BuildVersion, response.Message.Version)
		require.Equal(t, model.BuildCommit, response.Message.Commit)
		require.Equal(t, model.BuildDate, response.Message.Date)
	})
}
