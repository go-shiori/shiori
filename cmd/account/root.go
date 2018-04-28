package account

import (
	dt "github.com/RadhiFadlillah/shiori/database"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cError = color.New(color.FgHiRed)
	cIndex = color.New(color.FgHiCyan)
)

// NewAccountCmd creates new command for managing account
func NewAccountCmd(db dt.Database) *cobra.Command {
	// Create handler
	hdl := cmdHandler{db: db}

	// Create sub command
	addCmd := &cobra.Command{
		Use:   "add username",
		Short: "Create new account",
		Args:  cobra.ExactArgs(1),
		Run:   hdl.addAccount,
	}

	printCmd := &cobra.Command{
		Use:     "print",
		Short:   "Print the saved accounts",
		Args:    cobra.NoArgs,
		Aliases: []string{"list", "ls"},
		Run:     hdl.printAccounts,
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [usernames]",
		Short: "Delete the saved accounts",
		Long: "Delete accounts. " +
			"Accepts space-separated list of usernames. " +
			"If no arguments, all records will be deleted.",
		Run: hdl.deleteAccounts,
	}

	// Set sub command flags
	printCmd.Flags().StringP("search", "s", "", "Search accounts by username")
	deleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL accounts")

	// Create final root command
	rootCmd := &cobra.Command{
		Use:   "account",
		Short: "Manage account for accessing web interface",
	}

	rootCmd.AddCommand(addCmd, printCmd, deleteCmd)
	return rootCmd
}
