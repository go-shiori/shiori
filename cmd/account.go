package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

var (
	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "Manage account for accessing web interface.",
	}

	addAccountCmd = &cobra.Command{
		Use:   "add username",
		Short: "Create new account.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]

			fmt.Println("Username: " + username)
			fmt.Print("Password: ")

			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				cError.Println(err)
				return
			}

			fmt.Println()
			err = addAccount(username, string(bytePassword))
			if err != nil {
				cError.Println(err)
				return
			}
		},
	}

	printAccountCmd = &cobra.Command{
		Use:   "print",
		Short: "Print the saved accounts.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			keyword, _ := cmd.Flags().GetString("search")
			err := printAccounts(keyword)
			if err != nil {
				cError.Println(err)
				return
			}
		},
	}

	deleteAccountCmd = &cobra.Command{
		Use:   "delete [usernames]",
		Short: "Delete the saved accounts.",
		Long: "Delete accounts. " +
			"Accepts space-separated list of usernames. " +
			"If no arguments, all records will be deleted.",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			skipConfirmation, _ := cmd.Flags().GetBool("yes")

			// If no arguments, confirm to user
			if len(args) == 0 && !skipConfirmation {
				confirmDelete := ""
				fmt.Print("Remove ALL accounts? (y/n): ")
				fmt.Scanln(&confirmDelete)

				if confirmDelete != "y" {
					fmt.Println("No accounts deleted")
					return
				}
			}

			err := DB.DeleteAccounts(args...)
			if err != nil {
				cError.Println(err)
				return
			}

			fmt.Println("Accounts has been deleted")
		},
	}
)

func init() {
	// Create flags
	printAccountCmd.Flags().StringP("search", "s", "", "Search accounts by username")
	deleteAccountCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL accounts")

	accountCmd.AddCommand(addAccountCmd)
	accountCmd.AddCommand(printAccountCmd)
	accountCmd.AddCommand(deleteAccountCmd)
	rootCmd.AddCommand(accountCmd)
}

func addAccount(username, password string) error {
	if username == "" {
		return fmt.Errorf("Username must not empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("Password must be at least 8 characters")
	}

	err := DB.CreateAccount(username, password)
	if err != nil {
		return err
	}

	return nil
}

func printAccounts(keyword string) error {
	accounts, err := DB.GetAccounts(keyword)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		cIndex.Print("- ")
		fmt.Println(account.Username)
	}

	return nil
}
