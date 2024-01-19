package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/stackus/hamlet"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hamlet",
	Short: "A templating language for Go",
	Long: `Hamlet is a templating language for Go. It's designed to be simple and easy to use.
It combines Go and Haml to create a powerful templating language that's easy to learn.`,
	Version: hamlet.Version(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
