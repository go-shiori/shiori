package cmd

import (
	"github.com/spf13/cobra"
)

func pocketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pocket source-file",
		Short: "Import bookmarks from Pocket's exported HTML file",
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}
