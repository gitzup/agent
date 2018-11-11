package cmd

import (
	"context"
	"io/ioutil"

	. "github.com/gitzup/agent/internal/logger"
	"github.com/gitzup/agent/pkg/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Process a build request.",
	Long:  `This command will build the provided build request.`,
	Run: func(cmd *cobra.Command, args []string) {
		Logger().Info(args)
		if len(args) < 1 {
			Logger().Fatal("build ID is required")
		}
		if len(args) < 2 {
			Logger().Fatal("build file is required (use '-' for stdin)")
		}

		id := args[0]
		pipelineFile := args[1]

		// TODO: handle pipeline file equaling "-"
		bytes, err := ioutil.ReadFile(pipelineFile)
		if err != nil {
			Logger().WithError(err).Fatalf("failed reading '%s'", pipelineFile)
		}

		request, err := build.New(id, workspacePath, bytes)
		if err != nil {
			Logger().WithError(err).Fatal("failed creating build request")
		}

		// TODO: support timeout by using "context.WithTimeout(..)" as the context to "request.Apply(ctx)" method
		err = request.Apply(context.WithValue(context.Background(), "request", request.Id()))
		if err != nil {
			Logger().WithError(err).Fatal("failed applying build request")
		}

		// TODO: receive apply result and print it as text/json
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
