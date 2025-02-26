package domains_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestDirExists(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("foo", 0755)

	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	domain := domains.NewStorageDomain(
		deps,
		fs,
	)

	require.True(t, domain.DirExists("foo"))
	require.False(t, domain.DirExists("foo/file"))
	require.False(t, domain.DirExists("bar"))
}

func TestFileExists(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("foo", 0755)
	fs.Create("foo/file")

	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	domain := domains.NewStorageDomain(
		deps,
		fs,
	)

	require.True(t, domain.FileExists("foo/file"))
	require.False(t, domain.FileExists("bar"))
}

func TestWriteFile(t *testing.T) {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	domain := domains.NewStorageDomain(
		deps,
		fs,
	)

	err := domain.WriteData("foo/file.ext", []byte("foo"))
	require.NoError(t, err)
	require.True(t, domain.FileExists("foo/file.ext"))
	require.True(t, domain.DirExists("foo"))
	handler, err := domain.FS().Open("foo/file.ext")
	require.NoError(t, err)
	defer handler.Close()

	data, err := afero.ReadAll(handler)
	require.NoError(t, err)

	require.Equal(t, "foo", string(data))
}

func TestSaveFile(t *testing.T) {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.TODO(), logger)

	domain := domains.NewStorageDomain(
		deps,
		fs,
	)

	tempFile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("foo")
	require.NoError(t, err)

	err = domain.WriteFile("foo/file.ext", tempFile)
	require.NoError(t, err)
	require.True(t, domain.FileExists("foo/file.ext"))
	require.True(t, domain.DirExists("foo"))
	handler, err := domain.FS().Open("foo/file.ext")
	require.NoError(t, err)
	defer handler.Close()

	data, err := afero.ReadAll(handler)
	require.NoError(t, err)

	require.Equal(t, "foo", string(data))
}
