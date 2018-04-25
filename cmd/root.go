package cmd

import (
	"fmt"
	"github.com/RadhiFadlillah/shiori/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// DB is database that used by this cli
	DB database.Database

	rootCmd = &cobra.Command{
		Use:   "shiori",
		Short: "Simple command-line bookmark manager built with Go",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir())
	viper.ReadInConfig()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
