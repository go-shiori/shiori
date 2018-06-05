package cmd

import (
	"github.com/RadhiFadlillah/shiori/cmd/account"
	"github.com/RadhiFadlillah/shiori/cmd/serve"
	dt "github.com/RadhiFadlillah/shiori/database"
	"github.com/spf13/cobra"
)

// NewShioriCmd creates new command for shiori
func NewShioriCmd(db dt.Database, dataDir string) *cobra.Command {
	// Create handler
	hdl := cmdHandler{
		db:      db,
		dataDir: dataDir,
	}

	// Create sub command
	addCmd := &cobra.Command{
		Use:   "add url",
		Short: "Bookmark the specified URL",
		Args:  cobra.ExactArgs(1),
		Run:   hdl.addBookmark,
	}

	printCmd := &cobra.Command{
		Use:   "print [indices]",
		Short: "Print the saved bookmarks",
		Long: "Show the saved bookmarks by its DB index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records with actual index from DB are shown.",
		Aliases: []string{"list", "ls"},
		Run:     hdl.printBookmarks,
	}

	searchCmd := &cobra.Command{
		Use:   "search keyword",
		Short: "Search bookmarks by submitted keyword",
		Long: "Search bookmarks by looking for matching keyword in bookmark's title and content. " +
			"If no keyword submitted, print all saved bookmarks. " +
			"Search results will be different depending on DBMS that used by shiori :\n" +
			"- sqlite3, search works using fts4 method: https://www.sqlite.org/fts3.html.\n" +
			"- mysql or mariadb, search works using natural language mode: https://dev.mysql.com/doc/refman/5.5/en/fulltext-natural-language.html.",
		Args: cobra.MaximumNArgs(1),
		Run:  hdl.searchBookmarks,
	}

	updateCmd := &cobra.Command{
		Use:   "update [indices]",
		Short: "Update the saved bookmarks",
		Long: "Update fields of an existing bookmark. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be updated. Update works differently depending on the flags:\n" +
			"- If indices are passed without any flags (--url, --title, --tag and --excerpt), read the URLs from DB and update titles from web.\n" +
			"- If --url is passed (and --title is omitted), update the title from web using the URL. While using this flag, update only accept EXACTLY one index.\n" +
			"While updating bookmark's tags, you can use - to remove tag (e.g. -nature to remove nature tag from this bookmark).",
		Run: hdl.updateBookmarks,
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [indices]",
		Short: "Delete the saved bookmarks",
		Long: "Delete bookmarks. " +
			"When a record is deleted, the last record is moved to the removed index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records will be deleted.",
		Run: hdl.deleteBookmarks,
	}

	openCmd := &cobra.Command{
		Use:   "open [indices]",
		Short: "Open the saved bookmarks",
		Long: "Open bookmarks in browser. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be opened.",
		Run: hdl.openBookmarks,
	}

	importCmd := &cobra.Command{
		Use:   "import source-file",
		Short: "Import bookmarks from HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
		Run:   hdl.importBookmarks,
	}

	exportCmd := &cobra.Command{
		Use:   "export target-file",
		Short: "Export bookmarks into HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
		Run:   hdl.exportBookmarks,
	}

	pocketCmd := &cobra.Command{
		Use:   "pocket source-file",
		Short: "Import bookmarks from Pocket's exported HTML file",
		Args:  cobra.ExactArgs(1),
		Run:   hdl.importPockets,
	}

	// Create sub command that has its own sub command
	accountCmd := account.NewAccountCmd(db)
	serveCmd := serve.NewServeCmd(db, dataDir)

	// Set sub command flags
	addCmd.Flags().StringP("title", "i", "", "Custom title for this bookmark.")
	addCmd.Flags().StringP("excerpt", "e", "", "Custom excerpt for this bookmark.")
	addCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	addCmd.Flags().BoolP("offline", "o", false, "Save bookmark without fetching data from internet.")

	printCmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	printCmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")

	searchCmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	searchCmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")
	searchCmd.Flags().StringSliceP("tags", "t", []string{}, "Search bookmarks with specified tag(s)")

	updateCmd.Flags().StringP("url", "u", "", "New URL for this bookmark.")
	updateCmd.Flags().StringP("title", "i", "", "New title for this bookmark.")
	updateCmd.Flags().StringP("excerpt", "e", "", "New excerpt for this bookmark.")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	updateCmd.Flags().BoolP("offline", "o", false, "Update bookmark without fetching data from internet.")
	updateCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and update ALL bookmarks")
	updateCmd.Flags().Bool("dont-overwrite", false, "Don't overwrite existing metadata. Useful when only want to update bookmark's content.")

	deleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL bookmarks")

	openCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and open ALL bookmarks")
	openCmd.Flags().BoolP("cache", "c", false, "Open the bookmark's cache in text-only mode")
	openCmd.Flags().Bool("trim-space", false, "Trim all spaces and newlines from the bookmark's cache")

	importCmd.Flags().BoolP("generate-tag", "t", false, "Auto generate tag from bookmark's category")

	// Create final root command
	rootCmd := &cobra.Command{
		Use:   "shiori",
		Short: "Simple command-line bookmark manager built with Go",
	}

	rootCmd.AddCommand(accountCmd, serveCmd, addCmd, printCmd, searchCmd,
		updateCmd, deleteCmd, openCmd, importCmd, exportCmd, pocketCmd)
	return rootCmd
}
