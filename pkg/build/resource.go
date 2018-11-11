package build

import (
	"context"
	. "github.com/gitzup/agent/internal/logger"
	"github.com/gitzup/agent/pkg/assets"
	"github.com/go-errors/errors"
)

// Represents a single resource in a build request.
type Resource interface {
	Request() Request
	Name() string
	Type() string
	Config() interface{}
	ConfigSchema() *assets.Schema
	WorkspacePath() string
	Init(ctx context.Context) error
	DiscoverState(ctx context.Context) error
	Apply(ctx context.Context) error
}

type resourceImpl struct {
	request         Request
	name            string
	resourceType    string
	resourceConfig  interface{}
	workspacePath   string
	configSchema    *assets.Schema
	initAction      Action
	discoveryAction Action
}

type resourceInitRequest struct {
	Id       string   `json:"id"`
	Resource Resource `json:"resource"`
}

type resourceInitResponse struct {
	ConfigSchema interface{} `json:"configSchema"`
	StateAction  struct {
		Image      string   `json:"image"`
		Entrypoint []string `json:"entrypoint"`
		Cmd        []string `json:"cmd"`
	} `json:"stateAction"`
}

func (res *resourceImpl) Request() Request {
	return res.request
}

func (res *resourceImpl) Name() string {
	return res.name
}

func (res *resourceImpl) Type() string {
	return res.resourceType
}

func (res *resourceImpl) Config() interface{} {
	return res.resourceConfig
}

func (res *resourceImpl) ConfigSchema() *assets.Schema {
	return res.configSchema
}

func (res *resourceImpl) WorkspacePath() string {
	return res.workspacePath
}

func (res *resourceImpl) Init(ctx context.Context) error {
	ctx = context.WithValue(ctx, "resource", res.Name())

	From(ctx).Info("Initializing resource")

	// initialize the resource
	var response resourceInitResponse
	err := res.initAction.Invoke(
		ctx,
		&resourceInitRequest{
			Id:       res.Request().Id(),
			Resource: res,
		},
		assets.GetInitResponseSchema(),
		&response,
	)
	if err != nil {
		return errors.WrapPrefix(err, "failed initializing resource", 0)
	}

	// build the resource configuration schema
	resourceConfigSchema, err := assets.New(
		response.ConfigSchema,
		assets.GetActionSchema(),
		assets.GetResourceSchema(),
		assets.GetBuildRequestSchema(),
		assets.GetBuildResponseSchema(),
		assets.GetInitRequestSchema(),
		assets.GetInitResponseSchema(),
	)
	if err != nil {
		return err
	}
	res.configSchema = resourceConfigSchema

	// use the configuration schema to validate the resource's configuration
	if err := res.configSchema.Validate(res.Config()); err != nil {
		return err
	}

	// read and set the resource's state discovery action
	res.discoveryAction = &actionImpl{
		resource:   res,
		name:       "state",
		image:      response.StateAction.Image,
		entrypoint: response.StateAction.Entrypoint,
		cmd:        response.StateAction.Cmd,
	}

	return nil
}

func (res *resourceImpl) DiscoverState(ctx context.Context) error {
	ctx = context.WithValue(ctx, "resource", res.Name())

	From(ctx).Info("Discovering state")

	// TODO eric: implement
	panic("not implemented")
}

func (res *resourceImpl) Apply(ctx context.Context) error {
	ctx = context.WithValue(ctx, "resource", res.Name())

	From(ctx).Info("Applying")

	// TODO eric: implement
	panic("not implemented")
}
