package domains

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

// ExternalArchiver implements the ArchiverDomain interface using external shell commands
type ExternalArchiver struct {
	deps *dependencies.Dependencies
}

func (d *ExternalArchiver) ArchiveBookmark(book *model.BookmarkDTO, logEnabled bool) error {
	archiveCmd := os.Getenv("SHIORI_EXTERNAL_ARCHIVER_ARCHIVE_COMMAND")
	if archiveCmd == "" {
		return fmt.Errorf("SHIORI_EXTERNAL_ARCHIVER_ARCHIVE_COMMAND environment variable is not set")
	}

	// Replace placeholders in the command
	archiveCmd = strings.ReplaceAll(archiveCmd, "{URL}", book.URL)
	archiveCmd = strings.ReplaceAll(archiveCmd, "{ID}", fmt.Sprintf("%d", book.ID))
	archiveCmd = strings.ReplaceAll(archiveCmd, "{DATA_DIR}", d.deps.Config().Storage.DataDir)

	// Execute the command
	cmd := exec.Command("sh", "-c", archiveCmd)
	if logEnabled {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute archive command: %v", err)
	}

	return nil
}

func (d *ExternalArchiver) HasArchive(book *model.BookmarkDTO) bool {
	hasCmd := os.Getenv("SHIORI_EXTERNAL_ARCHIVER_HAS_COMMAND")
	if hasCmd == "" {
		return false
	}

	// Replace placeholders in the command
	hasCmd = strings.ReplaceAll(hasCmd, "{ID}", fmt.Sprintf("%d", book.ID))
	hasCmd = strings.ReplaceAll(hasCmd, "{DATA_DIR}", d.deps.Config().Storage.DataDir)

	// Execute the command
	cmd := exec.Command("sh", "-c", hasCmd)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if the output indicates the archive exists
	return strings.TrimSpace(string(output)) == "true"
}

func (d *ExternalArchiver) GetBookmarkArchive(book *model.BookmarkDTO) (*warc.Archive, error) {
	getCmd := os.Getenv("SHIORI_EXTERNAL_ARCHIVER_GET_COMMAND")
	if getCmd == "" {
		return nil, fmt.Errorf("SHIORI_EXTERNAL_ARCHIVER_GET_COMMAND environment variable is not set")
	}

	// Replace placeholders in the command
	getCmd = strings.ReplaceAll(getCmd, "{ID}", fmt.Sprintf("%d", book.ID))
	getCmd = strings.ReplaceAll(getCmd, "{DATA_DIR}", d.deps.Config().Storage.DataDir)

	// Execute the command to get the archive path
	cmd := exec.Command("sh", "-c", getCmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute get command: %v", err)
	}

	archivePath := strings.TrimSpace(string(output))
	if archivePath == "" {
		return nil, fmt.Errorf("archive path is empty")
	}

	// Open the archive file
	return warc.Open(archivePath)
}

func NewExternalArchiver(deps *dependencies.Dependencies) *ExternalArchiver {
	return &ExternalArchiver{
		deps: deps,
	}
}
