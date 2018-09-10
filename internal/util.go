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
	"strings"
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

type DockerPullStatus struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error"`
	ProgressText   string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
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

	pullEvents := json.NewDecoder(reader)
	var event *DockerPullStatus
	var progressMap = make(map[string]float32, 10)
	for {
		if err := pullEvents.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return &DockerError{Msg: "could not pull image", Cause: err, Request: *request}
		}

		progressMap[event.ID+":"+event.Status] = float32(event.ProgressDetail.Current) / float32(event.ProgressDetail.Total) * 100
		switch event.Status {
		case "Downloading":
			progressMap[event.ID+":Pulling fs layer"] = 100
		case "Download complete":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
		case "Extracting":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
			progressMap[event.ID+":Verifying Checksum"] = 100
			progressMap[event.ID+":Download complete"] = 100
		case "Pull complete":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
			progressMap[event.ID+":Verifying Checksum"] = 100
			progressMap[event.ID+":Download complete"] = 100
			progressMap[event.ID+":Extracting"] = 100
			progressMap[event.ID+":Pull complete"] = 100
		case "Digest":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
			progressMap[event.ID+":Verifying Checksum"] = 100
			progressMap[event.ID+":Download complete"] = 100
			progressMap[event.ID+":Extracting"] = 100
			progressMap[event.ID+":Pull complete"] = 100
		}
		builder := strings.Builder{}
		for _, progress := range progressMap {
			if progress > 0 && progress < 100 {
				builder.WriteString(fmt.Sprintf("%d%% ", uint8(progress)))
			}
		}
		if builder.Len() > 0 {
			fmt.Print("\r" + builder.String() + "\033[K")
		}
	}
	fmt.Print("\r\033[K\r")
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
