package cmd

import (
	. "github.com/gitzup/agent/internal/logger"
	"github.com/spf13/cobra"
)

// Workspace to place all build request workspaces in
var workspacePath string

// Log output format; can be "auto", "json", "plain" or "pretty":
//  * "auto": if a TTY is attached, acts the same as "pretty"; otherwise uses "json"
//  * "json": each log entry will be a JSON object containing all available information such as msg, timestamp, etc
//  * "plain": human-friendly output (unlike JSON) but without ANSI colors
//  * "pretty": human-friendly output with ANSI colors
var logFormat string

// Minimum log level to accept for output. Any log statements with a lower level will not be printed. Can be:
//  * trace
//  * debug
//  * info
//  * warn
//  * error
var logLevel string

// Whether to include caller information for each log entry. This has significant performance overhead and thus should
// only be used in debugging sessions or local development.
var caller bool

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:     "agent",
	Version: "1.0.0-alpha.1",
	Short:   "Gitzup agent executes pipelines",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitLogger(cmd.Root().Version, caller, logLevel, logFormat)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&workspacePath, "workspace", "w", ".", "Workspace location")
	rootCmd.PersistentFlags().StringVar(&logFormat, "logformat", "auto", "Log output format (auto, json, plain, pretty)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().BoolVarP(&caller, "caller", "c", false, "Include caller information in log output")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		Logger().WithError(err).Fatal("Execution error")
	}
}
