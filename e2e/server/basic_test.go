package e2e

import (
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/stretchr/testify/require"
)

func TestServerBasic(t *testing.T) {
	container := e2eutil.NewShioriContainer(t, "")

	t.Run("liveness endpoint", func(t *testing.T) {
		req, err := http.Get("http://localhost:" + container.GetPort() + "/system/liveness")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, req.StatusCode)
	})
}
