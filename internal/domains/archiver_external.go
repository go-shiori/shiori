package domains

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
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

	// Log the command being executed
	logger := d.deps.Logger()
	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
		"url":        book.URL,
		"command":    archiveCmd,
	}).Debug("External archiver: executing archive command")

	// Execute the command
	cmd := exec.Command("sh", "-c", archiveCmd)
	
	// Capture stderr for logging
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	
	if logEnabled {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		stderrOutput := strings.TrimSpace(stderrBuf.String())
		if stderrOutput != "" {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
				"stderr":     stderrOutput,
			}).Error("External archiver: archive command failed")
		} else {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
			}).Error("External archiver: archive command failed")
		}
		return fmt.Errorf("failed to execute archive command: %v", err)
	}

	// Log stderr output if there was any
	stderrOutput := strings.TrimSpace(stderrBuf.String())
	if stderrOutput != "" {
		logger.WithFields(logrus.Fields{
			"bookmark_id": book.ID,
			"stderr":     stderrOutput,
		}).Debug("External archiver: archive command stderr output")
	}

	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
	}).Debug("External archiver: successfully archived bookmark")

	return nil
}

func (d *ExternalArchiver) HasArchive(book *model.BookmarkDTO) bool {
	hasCmd := os.Getenv("SHIORI_EXTERNAL_ARCHIVER_HAS_COMMAND")
	if hasCmd == "" {
		return false
	}

	// Replace placeholders in the command
	hasCmd = strings.ReplaceAll(hasCmd, "{ID}", fmt.Sprintf("%d", book.ID))

	// Log the command being executed
	logger := d.deps.Logger()
	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
		"command":    hasCmd,
	}).Debug("External archiver: executing has archive command")

	// Execute the command
	cmd := exec.Command("sh", "-c", hasCmd)
	
	// Capture stderr for logging
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	
	output, err := cmd.Output()
	if err != nil {
		stderrOutput := strings.TrimSpace(stderrBuf.String())
		if stderrOutput != "" {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
				"stderr":     stderrOutput,
			}).Debug("External archiver: has archive command failed")
		} else {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
			}).Debug("External archiver: has archive command failed")
		}
		return false
	}

	// Log stderr output if there was any
	stderrOutput := strings.TrimSpace(stderrBuf.String())
	if stderrOutput != "" {
		logger.WithFields(logrus.Fields{
			"bookmark_id": book.ID,
			"stderr":     stderrOutput,
		}).Debug("External archiver: has archive command stderr output")
	}

	hasArchive := strings.TrimSpace(string(output)) == "true"
	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
		"has_archive": hasArchive,
	}).Debug("External archiver: has archive result")

	return hasArchive
}

func (d *ExternalArchiver) GetBookmarkArchive(book *model.BookmarkDTO) (model.Archive, error) {
	getCmd := os.Getenv("SHIORI_EXTERNAL_ARCHIVER_GET_COMMAND")
	if getCmd == "" {
		return nil, fmt.Errorf("SHIORI_EXTERNAL_ARCHIVER_GET_COMMAND environment variable is not set")
	}

	// Replace placeholders in the command
	getCmd = strings.ReplaceAll(getCmd, "{ID}", fmt.Sprintf("%d", book.ID))

	// Log the command being executed
	logger := d.deps.Logger()
	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
		"command":    getCmd,
	}).Debug("External archiver: executing get archive command")

	// Execute the command to get the archive contents
	cmd := exec.Command("sh", "-c", getCmd)
	
	// Capture stderr for logging
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	
	output, err := cmd.Output()
	if err != nil {
		stderrOutput := strings.TrimSpace(stderrBuf.String())
		if stderrOutput != "" {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
				"stderr":     stderrOutput,
			}).Error("External archiver: get archive command failed")
		} else {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
			}).Error("External archiver: get archive command failed")
		}
		return nil, fmt.Errorf("failed to execute get command: %v", err)
	}

	// Log stderr output if there was any
	stderrOutput := strings.TrimSpace(stderrBuf.String())
	if stderrOutput != "" {
		logger.WithFields(logrus.Fields{
			"bookmark_id": book.ID,
			"stderr":     stderrOutput,
		}).Debug("External archiver: get archive command stderr output")
	}

	// Check if the output is empty
	archiveContent := strings.TrimSpace(string(output))
	if archiveContent == "" {
		logger.WithFields(logrus.Fields{
			"bookmark_id": book.ID,
		}).Error("External archiver: archive content is empty")
		return nil, fmt.Errorf("archive content is empty")
	}

	// Log successful retrieval
	logger.WithFields(logrus.Fields{
		"bookmark_id":     book.ID,
		"content_length": len(archiveContent),
	}).Debug("External archiver: successfully retrieved archive content")

	// Create a SingleFileArchive with the content
	return model.NewSingleFileArchive([]byte(archiveContent)), nil
}

func (d *ExternalArchiver) DeleteArchive(book *model.BookmarkDTO) error {
	deleteCmd := os.Getenv("SHIORI_EXTERNAL_ARCHIVER_DELETE_COMMAND")
	if deleteCmd == "" {
		// If no delete command is configured, try to delete from /tmp/tmp-shiori
		archivePath := fmt.Sprintf("/tmp/tmp-shiori/archive_%d", book.ID)
		if err := os.Remove(archivePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete external archive: %v", err)
		}
		return nil
	}

	// Replace placeholders in the command
	deleteCmd = strings.ReplaceAll(deleteCmd, "{ID}", fmt.Sprintf("%d", book.ID))

	// Log the command being executed
	logger := d.deps.Logger()
	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
		"command":    deleteCmd,
	}).Debug("External archiver: executing delete archive command")

	// Execute the command
	cmd := exec.Command("sh", "-c", deleteCmd)
	
	// Capture stderr for logging
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	
	if err := cmd.Run(); err != nil {
		stderrOutput := strings.TrimSpace(stderrBuf.String())
		if stderrOutput != "" {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
				"stderr":     stderrOutput,
			}).Error("External archiver: delete archive command failed")
		} else {
			logger.WithFields(logrus.Fields{
				"bookmark_id": book.ID,
				"error":      err,
			}).Error("External archiver: delete archive command failed")
		}
		return fmt.Errorf("failed to execute delete command: %v", err)
	}

	// Log stderr output if there was any
	stderrOutput := strings.TrimSpace(stderrBuf.String())
	if stderrOutput != "" {
		logger.WithFields(logrus.Fields{
			"bookmark_id": book.ID,
			"stderr":     stderrOutput,
		}).Debug("External archiver: delete archive command stderr output")
	}

	logger.WithFields(logrus.Fields{
		"bookmark_id": book.ID,
	}).Debug("External archiver: successfully deleted archive")

	return nil
}

func NewExternalArchiver(deps *dependencies.Dependencies) *ExternalArchiver {
	return &ExternalArchiver{
		deps: deps,
	}
}
