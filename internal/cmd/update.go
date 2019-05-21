package cmd

import "github.com/spf13/cobra"

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [indices]",
		Short: "Update the saved bookmarks",
		Long: "Update fields of an existing bookmark. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be updated. Update works differently depending on the flags:\n" +
			"- If indices are passed without any flags (--url, --title, --tag and --excerpt), read the URLs from DB and update titles from web.\n" +
			"- If --url is passed (and --title is omitted), update the title from web using the URL. While using this flag, update only accept EXACTLY one index.\n" +
			"While updating bookmark's tags, you can use - to remove tag (e.g. -nature to remove nature tag from this bookmark).",
	}

	cmd.Flags().StringP("url", "u", "", "New URL for this bookmark.")
	cmd.Flags().StringP("title", "i", "", "New title for this bookmark.")
	cmd.Flags().StringP("excerpt", "e", "", "New excerpt for this bookmark.")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	cmd.Flags().BoolP("offline", "o", false, "Update bookmark without fetching data from internet.")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and update ALL bookmarks")
	cmd.Flags().Bool("dont-overwrite", false, "Don't overwrite existing metadata. Useful when only want to update bookmark's content.")

	return cmd
}
