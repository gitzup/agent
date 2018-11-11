package build

import (
	"context"
	"fmt"
	"github.com/gitzup/agent/internal/docker"
	. "github.com/gitzup/agent/internal/logger"
	"github.com/gitzup/agent/pkg/assets"
	"github.com/go-errors/errors"
	"time"
)

type Action interface {
	Resource() Resource
	Name() string
	Image() string
	Entrypoint() []string
	Cmd() []string
	Invoke(ctx context.Context, input interface{}, outputSchema *assets.Schema, output interface{}) error
}

type actionImpl struct {
	resource   Resource
	name       string
	image      string
	entrypoint []string
	cmd        []string
}

func (act *actionImpl) Resource() Resource {
	return act.resource
}

func (act *actionImpl) Name() string {
	return act.name
}

func (act *actionImpl) Image() string {
	return act.image
}

func (act *actionImpl) Entrypoint() []string {
	return act.entrypoint
}

func (act *actionImpl) Cmd() []string {
	return act.cmd
}

func (act *actionImpl) Invoke(ctx context.Context, input interface{}, outputSchema *assets.Schema, output interface{}) error {
	ctx = context.WithValue(ctx, "resource", act.resource.Name())
	ctx = context.WithValue(ctx, "action", act.Name())

	From(ctx).Infof("Invoking action '%s'", act.Name())

	// TODO: parametrize "progress"
	err := docker.Pull(ctx, false, act.Image())
	if err != nil {
		return err
	}

	// container name
	containerName := fmt.Sprintf("%s-%s-%s", act.Resource().Request().Id(), act.Resource().Name(), act.Name())

	// create container environment
	env := []string{
		fmt.Sprintf("GITZUP=%t", true),
		fmt.Sprintf("GITZUP_RESOURCE_NAME=%s", act.Resource().Name()),
		fmt.Sprintf("GITZUP_RESOURCE_TYPE=%s", act.Resource().Type()),
		fmt.Sprintf("GITZUP_ACTION_NAME=%s", act.Name()),
	}

	// volumes
	volumes := map[string]struct{}{}

	// result handler
	handler := docker.CreateJsonResultParser("/gitzup/result.json", outputSchema, &output)

	// create a timeout context
	// TODO: support timeout provided by resource itself (eg. from state action)
	runCtx, runCtxCancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer runCtxCancelFunc()

	// execute Docker image for this action
	if err = docker.Run(runCtx, act.Image(), containerName, env, volumes, input, nil, handler); err != nil {
		return errors.WrapPrefix(err, fmt.Sprintf("action '%s' failed", act.Name()), 0)
	}

	return nil
}
