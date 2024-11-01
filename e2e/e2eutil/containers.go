package e2eutil

import (
	"context"
	"os"
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

var testContainersProviderType testcontainers.ProviderType = testcontainers.ProviderDocker

func init() {
	// If TESTCONTAINERS_PROVIDER is set to podman, use podman
	// NOTE: This is EXPERIMENTAL since there are some issues running the e2e tests using podman,
	// testcontainers implies that it supports podman but I couldn't make it run in my tests.
	// YMMV.
	// More info: https://golang.testcontainers.org/system_requirements/using_podman/
	if os.Getenv("TESTCONTAINERS_PROVIDER") == "podman" {
		testContainersProviderType = testcontainers.ProviderPodman
	}
}

func newBuildArg(value string) *string {
	return &value
}

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
		ProviderType: testContainersProviderType,
		ContainerRequest: testcontainers.ContainerRequest{
			Cmd:          []string{"server", "--log-level", "debug"},
			ExposedPorts: []string{shioriPort},
			WaitingFor:   wait.ForLog(shioriExpectedStartupMessage).WithStartupTimeout(shioriExpectedStartupSeconds * time.Second),
		},
		Started: true,
	}

	if tag != "" {
		containerDefinition.ContainerRequest.FromDockerfile = testcontainers.FromDockerfile{}
		containerDefinition.Image = "gchr.io/go-shiori/shiori:" + tag
	} else {
		containerDefinition.FromDockerfile = testcontainers.FromDockerfile{
			Context:    "../..",
			Dockerfile: "Dockerfile.e2e",
			KeepImage:  true,
			BuildArgs: map[string]*string{
				"ALPINE_VERSION": newBuildArg(os.Getenv("CONTAINER_ALPINE_VERSION")),
				"GOLANG_VERSION": newBuildArg(os.Getenv("GOLANG_VERSION")),
			},
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
