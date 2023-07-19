package config

import (
	"fmt"
	"os"
	"path/filepath"

	gap "github.com/muesli/go-app-paths"
)

func getStorageDirectory(portableMode bool) (string, error) {
	// If in portable mode, uses directory of executable
	if portableMode {
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}

		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "shiori-data"), nil
	}

	// Try to use platform specific app path
	userScope := gap.NewScope(gap.User, "shiori")
	dataDir, err := userScope.DataPath("")
	if err == nil {
		return dataDir, nil
	}

	return "", fmt.Errorf("couldn't determine the data directory")
}
