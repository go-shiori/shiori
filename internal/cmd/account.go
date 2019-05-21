package cmd

import (
	"github.com/spf13/cobra"
)

func accountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage account for accessing web interface",
	}

	cmd.AddCommand(
		accountAddCmd(),
		accountPrintCmd(),
		accountDeleteCmd(),
	)

	return cmd
}
