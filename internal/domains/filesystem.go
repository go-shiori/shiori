package domains

import (
	"os"

	"github.com/sirupsen/logrus"
)

type FilesystemDomain struct {
	logger  *logrus.Logger
	dataDir string
}

func (d *FilesystemDomain) FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func NewFilesystemDomain(logger *logrus.Logger, dataDir string) FilesystemDomain {
	return FilesystemDomain{
		logger:  logger,
		dataDir: dataDir,
	}
}
