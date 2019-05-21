package cmd

import (
	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add url",
		Short: "Bookmark the specified URL",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("title", "i", "", "Custom title for this bookmark.")
	cmd.Flags().StringP("excerpt", "e", "", "Custom excerpt for this bookmark.")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	cmd.Flags().BoolP("offline", "o", false, "Save bookmark without fetching data from internet.")

	return cmd
}
