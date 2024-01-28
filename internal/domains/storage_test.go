package domains_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/domains"
)

func TestDirExists(t *testing.T) {
	path, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	t.Log(path)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		path,
	)

	require.NoError(t, domain.MkDirAll("foo", os.ModePerm))

	require.True(t, domain.DirExists("foo"))
	require.False(t, domain.DirExists("foo/file"))
	require.False(t, domain.DirExists("bar"))
}

func TestFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		tmpDir,
	)

	require.NoError(t, domain.MkDirAll("foo", os.ModePerm))
	tmpFile, err := domain.Create("foo/file")
	require.NoError(t, err)
	tmpFile.Close()

	require.True(t, domain.FileExists("foo/file"))
	require.False(t, domain.FileExists("bar"))
}

func TestWriteData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		tmpDir,
	)

	err = domain.WriteData("foo/file.ext", []byte("foo"))
	require.NoError(t, err)
	require.True(t, domain.FileExists("foo/file.ext"))
	require.True(t, domain.DirExists("foo"))
	handler, err := domain.Open("foo/file.ext")
	require.NoError(t, err)
	defer handler.Close()

	data, err := io.ReadAll(handler)
	require.NoError(t, err)
	require.Equal(t, "foo", string(data))
}

func TestSaveFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		tmpDir,
	)

	tmpFile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer tmpFile.Close()

	tmpFile.Write([]byte("foo"))
	tmpFile.Seek(0, 0)

	require.NoError(t, domain.WriteFile("foo/file.ext", tmpFile))
	require.True(t, domain.FileExists("foo/file.ext"))
	require.True(t, domain.DirExists("foo"))
	handler, err := domain.Open("foo/file.ext")
	require.NoError(t, err)
	defer handler.Close()

	data, err := io.ReadAll(handler)
	require.NoError(t, err)

	require.Equal(t, "foo", string(data))
}

func TestRemoveFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	domain := domains.NewStorageDomain(
		&dependencies.Dependencies{
			Config: config.ParseServerConfiguration(context.TODO(), logrus.New()),
			Log:    logrus.New(),
		},
		tmpDir,
	)

	require.NoError(t, domain.WriteData("foo/file.ext", []byte("foo")))
	require.True(t, domain.FileExists("foo/file.ext"))

	require.NoError(t, domain.Remove("foo/file.ext"))
	require.False(t, domain.FileExists("foo/file.ext"))
}
