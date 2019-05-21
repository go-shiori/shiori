package cmd

import "github.com/spf13/cobra"

func accountPrintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "print",
		Short:   "Print the saved accounts",
		Args:    cobra.NoArgs,
		Aliases: []string{"list", "ls"},
	}

	cmd.Flags().StringP("search", "s", "", "Search accounts by username")

	return cmd
}
