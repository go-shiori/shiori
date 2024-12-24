package e2e

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/stretchr/testify/require"
)

func TestAuthLogin(t *testing.T) {
	container := e2eutil.NewShioriContainer(t, "")

	t.Run("login ok", func(t *testing.T) {
		req, err := http.Post(
			"http://localhost:"+container.GetPort()+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"username": "shiori", "password": "gopher"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, req.StatusCode)
	})

	t.Run("wrong credentials", func(t *testing.T) {
		req, err := http.Post(
			"http://localhost:"+container.GetPort()+"/api/v1/auth/login",
			"application/json",
			bytes.NewReader([]byte(`{"username": "wrong", "password": "wrong"}`)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, req.StatusCode)
	})
}
