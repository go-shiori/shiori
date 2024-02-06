package config

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHostnameVariable(t *testing.T) {
	os.Setenv("HOSTNAME", "test_hostname")
	defer os.Unsetenv("HOSTNAME")

	log := logrus.New()
	cfg := ParseServerConfiguration(context.TODO(), log)

	require.Equal(t, "test_hostname", cfg.Hostname)
}

// TestBackwardsCompatibility tests that the old environment variables changed from 1.5.5 onwards
// are still supported and working with the new configuration system.
func TestBackwardsCompatibility(t *testing.T) {
	for _, env := range []struct {
		env  string
		want string
		eval func(t *testing.T, cfg *Config)
	}{
		{"HOSTNAME", "test_hostname", func(t *testing.T, cfg *Config) {
			require.Equal(t, "test_hostname", cfg.Hostname)
		}},
		{"SHIORI_DIR", "test", func(t *testing.T, cfg *Config) {
			require.Equal(t, "test", cfg.Storage.DataDir)
		}},
		{"SHIORI_DBMS", "test", func(t *testing.T, cfg *Config) {
			require.Equal(t, "test", cfg.Database.DBMS)
		}},
	} {
		t.Run(env.env, func(t *testing.T) {
			os.Setenv(env.env, env.want)
			t.Cleanup(func() {
				os.Unsetenv(env.env)
			})

			log := logrus.New()
			cfg := ParseServerConfiguration(context.Background(), log)
			env.eval(t, cfg)
		})
	}
}

func TestReadDotEnv(t *testing.T) {
	log := logrus.New()

	for _, testCase := range []struct {
		name string
		line string
		env  map[string]string
	}{
		{"empty", "", map[string]string{}},
		{"comment", "# comment", map[string]string{}},
		{"ignore invalid lines", "invalid line", map[string]string{}},
		{"single variable", "SHIORI_HTTP_PORT=9999", map[string]string{"SHIORI_HTTP_PORT": "9999"}},
		{"multiple variable", "SHIORI_HTTP_PORT=9999\nSHIORI_HTTP_SECRET_KEY=123123", map[string]string{"SHIORI_HTTP_PORT": "9999", "SHIORI_HTTP_SECRET_KEY": "123123"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))

			// Write the .env file in the temporary directory
			handler, err := os.OpenFile(".env", os.O_CREATE|os.O_WRONLY, 0655)
			require.NoError(t, err)
			handler.Write([]byte(testCase.line + "\n"))
			handler.Close()

			e := readDotEnv(log)

			require.Equal(t, testCase.env, e)
		})
	}

	t.Run("no file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "")
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))

		e := readDotEnv(log)

		require.Equal(t, map[string]string{}, e)
	})
}

func TestConfigSetDefaults(t *testing.T) {
	log := logrus.New()
	cfg := ParseServerConfiguration(context.TODO(), log)
	cfg.SetDefaults(log, false)

	require.NotEmpty(t, cfg.Http.SecretKey)
	require.NotEmpty(t, cfg.Storage.DataDir)
	require.NotEmpty(t, cfg.Database.URL)
}
