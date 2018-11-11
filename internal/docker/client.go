package docker

import (
	"github.com/docker/docker/client"
)

// Docker Client
var cli *client.Client

// Initialize the package
func init() {
	dockerCli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	cli = dockerCli
}
