package domains_test

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/psanford/memfs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestDirExists(t *testing.T) {
	fs := memfs.New()
	fs.MkdirAll("foo", 0755)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		fs,
	)

	require.True(t, domain.DirExists("foo"))
	require.False(t, domain.DirExists("foo/file"))
	require.False(t, domain.DirExists("bar"))
}

func TestFileExists(t *testing.T) {
	fs := memfs.New()
	fs.MkdirAll("foo", 0755)
	fs.WriteFile("foo/file", []byte("hello world"), 0644)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		fs,
	)

	require.True(t, domain.FileExists("foo/file"))
	require.False(t, domain.FileExists("bar"))
}
