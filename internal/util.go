package internal

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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

type DockerRequest struct {
	Image           string
	ContainerName   string
	Env             []string
	Volumes         map[string]struct{}
	Input           interface{}
	PreExitHandler  func(c container.ContainerCreateCreatedBody, request *DockerRequest) error
	PostExitHandler func(c container.ContainerCreateCreatedBody, request *DockerRequest) error
}

type DockerError struct {
	error
	Msg     string
	Cause   error
	Request DockerRequest
}

func (e *DockerError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", e.Msg, e.Cause.Error(), e.Request.ContainerName)
}

func (request *DockerRequest) ensureImagePresent() error {
	ctx := context.Background()
	defer ctx.Done()

	// list images
	imageListArgs := filters.NewArgs()
	imageListArgs.Add("reference", request.Image)
	imageList, err := cli.ImageList(ctx, types.ImageListOptions{Filters: imageListArgs})
	if err != nil {
		return &DockerError{Msg: "could not fetch local image list", Cause: err, Request: *request}
	}

	// if it is already present, return
	if len(imageList) > 0 {
		return nil
	}

	// pull it
	log.Printf("Pulling image '%s'...\n", request.Image)
	reader, err := cli.ImagePull(ctx, request.Image, types.ImagePullOptions{All: true})
	if err != nil {
		return &DockerError{Msg: "could not pull image", Cause: err, Request: *request}
	}
	defer reader.Close()

	// TODO: pretty-print image-pull reader messages to stderr
	if io.Copy(os.Stderr, reader); err != nil {
		return &DockerError{Msg: "could not stream image pull progress", Cause: err, Request: *request}
	}
	return nil
}

func (request *DockerRequest) InvokeAndDoWithContainer() error {
	request.ensureImagePresent()

	ctx := context.Background()
	defer ctx.Done()

	// create container
	c, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Domainname:   "gitzup.local",
			AttachStdin:  false,
			AttachStdout: true, // TODO: no need here
			AttachStderr: true, // TODO: no need here
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          request.Env,
			Image:        request.Image,
			Volumes:      request.Volumes,
		},
		&container.HostConfig{AutoRemove: false},
		nil,
		request.ContainerName)
	if err != nil {
		return &DockerError{Msg: "could not create container", Cause: err, Request: *request}
	}
	defer func() {
		if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{}); err != nil {
			log.Printf("Failed removing container '%s': %v", c.ID, err)
		}
	}()

	// start container
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return &DockerError{Msg: "could not start container", Cause: err, Request: *request}
	}

	// if input provided, attach to container and send the input to stdin
	if request.Input != nil {
		resp, err := cli.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{Stream: true, Stdin: true})
		if err != nil {
			return &DockerError{Msg: "could not attach to container", Cause: err, Request: *request}
		}
		defer resp.Close()

		// send input to the container's stdin
		b, err := json.Marshal(request.Input)
		if err != nil {
			return &DockerError{Msg: "could not send input to container", Cause: err, Request: *request}
		}
		resp.Conn.Write(b)
		resp.CloseWrite()
	}

	// if pre-exit handler provided, invoke it now
	if request.PreExitHandler != nil {
		// handle the container
		err = request.PreExitHandler(c, request)
		if err != nil {
			return &DockerError{Msg: "container pre-exit handler failed", Cause: err, Request: *request}
		}
	}

	// stream logs to our stdout
	stdout, err := cli.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Follow:     true,
	})
	defer stdout.Close()
	if err != nil {
		return &DockerError{Msg: "could not stream back container stdout", Cause: err, Request: *request}
	}
	stderr, err := cli.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: false,
		ShowStderr: true,
		Follow:     true,
	})
	defer stderr.Close()
	if err != nil {
		return &DockerError{Msg: "could not stream back container stderr", Cause: err, Request: *request}
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// wait for container to finish
	ctx30Sec, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := cli.ContainerWait(ctx30Sec, c.ID); err != nil {
		log.Println("Killing container (SIGTERM)")
		if err := cli.ContainerKill(ctx, c.ID, "SIGTERM"); err != nil {
			log.Println("Failed; killing container (SIGKILL)")
			if err := cli.ContainerKill(ctx, c.ID, "SIGKILL"); err != nil {
				log.Printf("Failed killing container '%s'", c.ID)
			}
		}
		time.Sleep(2 * time.Second)
		return &DockerError{Msg: fmt.Sprintf("container timed out (%s)", c.ID), Cause: err, Request: *request}
	}
	cntr, err := cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return &DockerError{Msg: "failed inspecting container", Cause: err, Request: *request}
	}
	exitCode := cntr.State.ExitCode
	if exitCode != 0 {
		return &DockerError{Msg: "container terminated with exit code " + string(exitCode), Cause: err, Request: *request}
	}

	// if post-exit handler provided, invoke it now
	if request.PostExitHandler != nil {
		// handle the container
		err = request.PostExitHandler(c, request)
		if err != nil {
			return &DockerError{Msg: "container post-exit handler failed", Cause: err, Request: *request}
		}
	}

	return nil
}

func CreateJsonResultParser(path string, schema *Schema, response interface{}) func(c container.ContainerCreateCreatedBody, request *DockerRequest) error {
	return func(c container.ContainerCreateCreatedBody, request *DockerRequest) error {

		ctx := context.Background()
		defer ctx.Done()

		// read JSON response from resource (fail if missing or invalid)
		reader, _, err := cli.CopyFromContainer(ctx, c.ID, path)
		if err != nil {
			return &DockerError{
				Msg:     fmt.Sprintf("protocol error: could not find '%s' in container", path),
				Cause:   err,
				Request: *request,
			}
		}
		defer reader.Close()
		tr := tar.NewReader(reader)
		_, err = tr.Next()
		if err == io.EOF {
			return &DockerError{
				Msg:     fmt.Sprintf("protocol error: could not find '%s' in container", path),
				Cause:   err,
				Request: *request,
			}
		} else if err != nil {
			return &DockerError{
				Msg:     fmt.Sprintf("failed reading '%s' in container", path),
				Cause:   err,
				Request: *request,
			}
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			return &DockerError{
				Msg:     fmt.Sprintf("failed reading '%s' in container", path),
				Cause:   err,
				Request: *request,
			}
		}
		err = schema.ParseAndValidate(response, b)
		if err != nil {
			return &DockerError{
				Msg:     fmt.Sprintf("protocol error: invalid content in '%s' in container", path),
				Cause:   err,
				Request: *request,
			}
		}

		return nil
	}
}
