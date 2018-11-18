package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	. "github.com/gitzup/agent/internal/logger"
	"github.com/go-errors/errors"
)

type ContainerRunHandler func(ctx context.Context, c container.ContainerCreateCreatedBody) error

func printLoop(ctx context.Context, input io.ReadCloser, printer func(logger *logrus.Entry, line string)) {
	in := bufio.NewScanner(input)
	for in.Scan() {
		// TODO: exit when context is canceled
		printer(From(ctx), in.Text())
		if err := in.Err(); err != nil {
			From(ctx).WithError(err).Error("failed copying output to logger")
		}
	}
}

func Run(
	ctx context.Context,
	image string,
	containerName string,
	env []string,
	volumes map[string]struct{},
	input interface{},
	preExitHandler ContainerRunHandler,
	postExitHandler ContainerRunHandler) *errors.Error {

	// work with a child context which has the "containerName" key
	ctx = context.WithValue(ctx, "container", containerName)

	// create container
	c, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Domainname:   "gitzup.local",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			StdinOnce:    true,
			Tty:          true,
			Env:          env,
			Image:        image,
			Volumes:      volumes,
		},
		&container.HostConfig{AutoRemove: false},
		nil,
		containerName)
	if err != nil {
		return errors.WrapPrefix(err, "failed creating container", 0)
	}
	defer func() {
		if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{}); err != nil {
			From(ctx).WithError(err).Warnf("failed removing container ID '%s'", c.ID)
		}
	}()

	// start container
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return errors.WrapPrefix(err, "failed starting container", 0)
	}

	// if input provided, attach to container and send the input to stdin
	if input != nil {
		resp, err := cli.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{Stream: true, Stdin: true})
		if err != nil {
			return errors.WrapPrefix(err, "failed attaching to container", 0)
		}
		defer resp.Close()

		// send input to the container's stdin
		b, err := json.Marshal(input)
		if err != nil {
			return errors.WrapPrefix(err, "failed serializing request input to JSON", 0)
		}
		_, err = resp.Conn.Write(b)
		if err != nil {
			return errors.WrapPrefix(err, "failed writing input to stdin of container", 0)
		}

		// I think this flushes the stdin into the container
		err = resp.CloseWrite()
		if err != nil {
			return errors.WrapPrefix(err, "failed writing input to stdin of container", 0)
		}
	}

	// stream logs to our stdout
	stdout, err := cli.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Follow:     true,
	})
	defer func() {
		err := stdout.Close()
		if err != nil {
			From(ctx).WithError(err).Warnf("failed closing stdout logs streams")
		}
	}()
	if err != nil {
		return errors.WrapPrefix(err, "failed fetching stdout log from container", 0)
	}
	go printLoop(ctx, stdout, func(logger *logrus.Entry, line string) { logger.Info(line) })

	// stream logs to our stderr
	stderr, err := cli.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: false,
		ShowStderr: true,
		Follow:     true,
	})
	defer func() {
		err := stderr.Close()
		if err != nil {
			From(ctx).WithError(err).Warnf("failed closing stderr logs streams")
		}
	}()
	if err != nil {
		return errors.WrapPrefix(err, "failed fetching stderr log from container", 0)
	}
	go printLoop(ctx, stderr, func(logger *logrus.Entry, line string) { logger.Warn(line) })

	// if pre-exit handler provided, invoke it now
	// TODO: run handler in goroutine, and abort when context is canceled
	if preExitHandler != nil {
		// handle the container
		err = preExitHandler(ctx, c)
		if err != nil {
			return errors.WrapPrefix(err, "failed invoking pre-exit callback on container", 0)
		}
	}

	// wait for container to finish
	ctx30Sec, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := cli.ContainerWait(ctx30Sec, c.ID); err != nil {
		From(ctx).WithError(err).Warnf("Killing container '%s' (SIGTERM)", c.ID)
		if err := cli.ContainerKill(ctx, c.ID, "SIGTERM"); err != nil {
			From(ctx).WithError(err).Warnf("Failed killing container '%s' using SIGTERM; will now use SIGKILL", c.ID)
			if err := cli.ContainerKill(ctx, c.ID, "SIGKILL"); err != nil {
				From(ctx).WithError(err).Errorf("Failed killing container '%s' using both SIGTERM and SIGKILL", c.ID)
			}
		}
		time.Sleep(2 * time.Second)
		return errors.WrapPrefix(err, "failed waiting for container", 0)
	}
	cntr, err := cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return errors.WrapPrefix(err, "failed inspecting container", 0)
	}
	exitCode := cntr.State.ExitCode
	if exitCode != 0 {
		return errors.WrapPrefix(err, fmt.Sprintf("container terminated with exit-code %d", exitCode), 0)
	}

	// if post-exit handler provided, invoke it now
	// TODO: run handler in goroutine, and abort when context is canceled
	if postExitHandler != nil {
		// handle the container
		err = postExitHandler(ctx, c)
		if err != nil {
			return errors.WrapPrefix(err, "failed invoking post-exit callback on container", 0)
		}
	}

	return nil
}
