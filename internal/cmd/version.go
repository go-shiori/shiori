package cmd

import (
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newVersionCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Output the shiori version",
		Run:   newVersionCommandHandler(logger),
	}

	return cmd
}

func newVersionCommandHandler(logger *logrus.Logger) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		cmd.Printf("Shiori version %s (build %s) at %s\n", model.Version, model.Commit, model.Date)
	}
}
