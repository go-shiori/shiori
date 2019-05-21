package cmd

import "github.com/spf13/cobra"

func printCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print [indices]",
		Short: "Print the saved bookmarks",
		Long: "Show the saved bookmarks by its DB index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records with actual index from database are shown.",
		Aliases: []string{"list", "ls"},
	}

	cmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	cmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")

	return cmd
}
