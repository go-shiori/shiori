package cmd

import (
	"github.com/spf13/cobra"
)

func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export target-file",
		Short: "Export bookmarks into HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}
