package cmd

import "github.com/spf13/cobra"

func accountDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [usernames]",
		Short: "Delete the saved accounts",
		Long: "Delete accounts. " +
			"Accepts space-separated list of usernames. " +
			"If no arguments, all records will be deleted.",
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL accounts")

	return cmd
}
