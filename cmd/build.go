package cmd

import (
	"io/ioutil"
	"log"

	"github.com/gitzup/agent/internal"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Process a build request.",
	Long:  `This command will build the providing build request.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		bytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			log.Fatal(err)
		}

		request, err := internal.ParseBuildRequest(id, bytes, workspacePath)
		if err != nil {
			log.Fatal(err)
		}

		err = request.Apply()
		if err != nil {
			log.Fatal(err)
		}

		// TODO: receive apply result and print it as text/json
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
