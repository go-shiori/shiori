package cmd

import "github.com/spf13/cobra"

func searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search keyword",
		Short: "Search bookmarks by submitted keyword",
		Long: "Search bookmarks by looking for matching keyword in bookmark's title and content. " +
			"If no keyword submitted, print all saved bookmarks. " +
			"Search results will be different depending on DBMS that used by shiori :\n" +
			"- sqlite3, search works using fts4 method: https://www.sqlite.org/fts3.html.\n" +
			"- mysql or mariadb, search works using natural language mode: https://dev.mysql.com/doc/refman/5.5/en/fulltext-natural-language.html.",
		Args: cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	cmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Search bookmarks with specified tag(s)")

	return cmd
}
