package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [indices]",
		Short: "Delete the saved bookmarks",
		Long: "Delete bookmarks. " +
			"When a record is deleted, the last record is moved to the removed index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL records will be deleted.",
		Aliases: []string{"rm"},
		Run:     deleteHandler,
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL bookmarks")

	return cmd
}

func deleteHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	// If no arguments (i.e all bookmarks going to be deleted), confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmDelete := ""
		fmt.Print("Remove ALL bookmarks? (y/N): ")
		fmt.Scanln(&confirmDelete)

		if confirmDelete != "y" {
			fmt.Println("No bookmarks deleted")
			return
		}
	}

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Printf("Failed to parse args: %v\n", err)
		os.Exit(1)
	}

	// Delete bookmarks from database
	err = db.DeleteBookmarks(cmd.Context(), ids...)
	if err != nil {
		cError.Printf("Failed to delete bookmarks: %v\n", err)
		os.Exit(1)
	}

	// Delete thumbnail image and archives from local disk
	if len(ids) == 0 {
		thumbDir := fp.Join(dataDir, "thumb")
		archiveDir := fp.Join(dataDir, "archive")
		os.RemoveAll(thumbDir)
		os.RemoveAll(archiveDir)
	} else {
		for _, id := range ids {
			strID := strconv.Itoa(id)
			imgPath := fp.Join(dataDir, "thumb", strID)
			archivePath := fp.Join(dataDir, "archive", strID)

			os.Remove(imgPath)
			os.Remove(archivePath)
		}
	}

	// Show finish message
	switch len(args) {
	case 0:
		fmt.Println("All bookmarks have been deleted")
	case 1, 2, 3, 4, 5:
		fmt.Printf("Bookmark(s) %s have been deleted\n", strings.Join(args, ", "))
	default:
		fmt.Println("Bookmark(s) have been deleted")
	}
}
