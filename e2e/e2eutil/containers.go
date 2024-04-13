package e2eutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	shioriPort                   = "8080/tcp"
	shioriExpectedStartupMessage = "started http server"
	shioriExpectedStartupSeconds = 5
)

type ShioriContainer struct {
	t *testing.T

	Container testcontainers.Container
}

func (sc *ShioriContainer) GetPort() string {
	mappedPort, err := sc.Container.MappedPort(context.Background(), shioriPort)
	require.NoError(sc.t, err)
	return mappedPort.Port()
}

// NewShioriContainer creates a new ShioriContainer which is a wrapper around a testcontainers.Container
// with some helpers for using while running Shiori E2E tests.
func NewShioriContainer(t *testing.T, tag string) ShioriContainer {
	containerDefinition := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Cmd:          []string{"server", "--log-level", "debug"},
			ExposedPorts: []string{shioriPort},
			WaitingFor:   wait.ForLog(shioriExpectedStartupMessage).WithStartupTimeout(shioriExpectedStartupSeconds * time.Second),
		},
		Started: true,
	}

	if tag != "" {
		containerDefinition.ContainerRequest.FromDockerfile = testcontainers.FromDockerfile{}
		containerDefinition.Image = "gchr.io/go-shiori/shiori:v" + tag
	} else {
		containerDefinition.FromDockerfile = testcontainers.FromDockerfile{
			Context:   ".",
			KeepImage: true,
		}
	}

	container, err := testcontainers.GenericContainer(context.Background(), containerDefinition)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	return ShioriContainer{
		t:         t,
		Container: container,
	}
}
