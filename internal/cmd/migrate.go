package cmd

import (
	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the database to the latest version",
		Run:   migrateHandler,
	}

	return cmd
}

func migrateHandler(cmd *cobra.Command, args []string) {
	if err := db.Migrate(); err != nil {
		cError.Printf("Error during migration: %s", err)
	}
}
