package cmd

import "github.com/spf13/cobra"

func accountAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add username",
		Short: "Create new account",
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}
