package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:     "agent",
	Version: "1.0.0-alpha.1",
	Short:   "Gitzup agent executes pipelines",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var workspacePath string

func init() {
	rootCmd.PersistentFlags().StringVarP(&workspacePath, "workspace", "w", ".", "Workspace location")
}
