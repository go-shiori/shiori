package cmd

import (
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Output the shiori version",
		Run:   newVersionCommandHandler(),
	}

	return cmd
}

func newVersionCommandHandler() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		cmd.Printf("Shiori version %s (build %s) at %s\n", model.BuildVersion, model.BuildCommit, model.BuildDate)
	}
}
