package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const version = "dev"

func TestServerBasic(t *testing.T) {
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ghcr.io/go-shiori/shiori:" + version,
			Cmd:          []string{"server", "--log-level", "debug"},
			ExposedPorts: []string{"8080/tcp"},
			HostConfigModifier: func(hc *container.HostConfig) {

			},
			WaitingFor: wait.ForLog("started http server").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	mappedPort, err := container.MappedPort(context.Background(), "8080/tcp")
	require.NoError(t, err)

	req, err := http.Get("http://localhost:" + mappedPort.Port() + "/system/liveness")
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, req.StatusCode)
}
