package account

import (
	"fmt"
	"syscall"

	dt "github.com/RadhiFadlillah/shiori/database"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// cmdHandler is handler for all action in AccountCmd
type cmdHandler struct {
	db dt.Database
}

// addAccount is handler for creating new account.
// Accept exactly one argument, i.e. username.
func (h *cmdHandler) addAccount(cmd *cobra.Command, args []string) {
	// Validate and show username
	username := args[0]
	if username == "" {
		cError.Println("Username must not be empty")
		return
	}

	fmt.Println("Username: " + username)

	// Read and validate password
	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		cError.Println(err)
		return
	}

	fmt.Println()
	strPassword := string(bytePassword)
	if len(strPassword) < 8 {
		cError.Println("Password must be at least 8 characters")
		return
	}

	// Save account to database
	err = h.db.CreateAccount(username, strPassword)
	if err != nil {
		cError.Println(err)
	}
}

// printAccounts is handler for showing all saved accounts.
// Can be used to search accounts by using flag -search.
func (h *cmdHandler) printAccounts(cmd *cobra.Command, args []string) {
	// Parse flags
	keyword, _ := cmd.Flags().GetString("search")

	// Fetch list accounts in database
	accounts, err := h.db.GetAccounts(keyword)
	if err != nil {
		cError.Println(err)
		return
	}

	// Show list accounts
	for _, account := range accounts {
		cIndex.Print("- ")
		fmt.Println(account.Username)
	}
}

// deleteAccounts is handler for deleting saved accounts.
func (h *cmdHandler) deleteAccounts(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	// If no arguments (i.e all accounts going to be deleted),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmDelete := ""
		fmt.Print("Remove ALL accounts? (y/n): ")
		fmt.Scanln(&confirmDelete)

		if confirmDelete != "y" {
			fmt.Println("No accounts deleted")
			return
		}
	}

	// Delete accounts in database
	err := h.db.DeleteAccounts(args...)
	if err != nil {
		cError.Println(err)
		return
	}

	fmt.Println("Account(s) have been deleted")
}
