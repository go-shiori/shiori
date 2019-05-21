package cmd

import (
	"github.com/spf13/cobra"
)

func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [indices]",
		Short: "Delete the saved bookmarks",
		Long: "Delete bookmarks. " +
			"When a record is deleted, the last record is moved to the removed index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL records will be deleted.",
		Aliases: []string{"rm"},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL bookmarks")

	return cmd
}
