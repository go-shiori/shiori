package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_printHandler_emptyJSON(t *testing.T) {
	if os.Getenv("SHIORI_TEST_EMPTY_JSON") == "1" {
		command := ShioriCmd()
		command.SetArgs([]string{"print", "--json"})
		if err := command.Execute(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	t.Setenv("SHIORI_DIR", t.TempDir())
	t.Setenv("SHIORI_DATABASE_URL", "")
	t.Setenv("SHIORI_DBMS", "")
	t.Setenv("SHIORI_HTTP_SECRET_KEY", "local-print-test-secret")
	t.Setenv("SHIORI_DEVELOPMENT", "false")

	executable, err := os.Executable()
	require.NoError(t, err)
	command := exec.Command(executable, "-test.run=^Test_printHandler_emptyJSON$")
	command.Env = append(os.Environ(), "SHIORI_TEST_EMPTY_JSON=1")
	output, err := command.Output()
	require.NoError(t, err)
	require.JSONEq(t, "[]", string(output))
}
