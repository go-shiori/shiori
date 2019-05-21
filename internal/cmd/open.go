package cmd

import (
	"github.com/spf13/cobra"
)

func openCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [indices]",
		Short: "Open the saved bookmarks",
		Long: "Open bookmarks in browser. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be opened.",
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and open ALL bookmarks")
	cmd.Flags().BoolP("cache", "c", false, "Open the bookmark's cache in text-only mode")
	cmd.Flags().Bool("trim-space", false, "Trim all spaces and newlines from the bookmark's cache")

	return cmd
}
