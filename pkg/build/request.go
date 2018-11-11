package build

import (
	"context"
	"github.com/go-errors/errors"
	"path"

	. "github.com/gitzup/agent/internal/logger"
	"github.com/gitzup/agent/pkg/assets"
)

// Represents a context for a single build request. Extends 'context.Context' and provides additional information and
// tools such as a tagged logger and workspace path.
type Request interface {
	Id() string
	Resources() map[string]Resource
	WorkspacePath() string
	Apply(ctx context.Context) error
}

type requestImpl struct {
	id            string
	resources     *map[string]*resourceImpl
	workspacePath string
}

func (req *requestImpl) Id() string {
	return req.id
}

func (req *requestImpl) Resources() map[string]Resource {
	resources := make(map[string]Resource)
	for name, pres := range *req.resources {
		resources[name] = pres
	}
	return resources
}

func (req *requestImpl) WorkspacePath() string {
	return req.workspacePath
}

func (req *requestImpl) Apply(ctx context.Context) error {
	From(ctx).Info("Applying build request")

	for _, resource := range req.Resources() {
		err := resource.Init(ctx)
		if err != nil {
			return err
		}
	}

	for _, resource := range req.Resources() {
		err := resource.DiscoverState(ctx)
		if err != nil {
			return err
		}
	}

	for _, resource := range req.Resources() {
		err := resource.Apply(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new build request context.
func New(id string, workspacePath string, b []byte) (req Request, err error) {

	// validate & parse the build request
	var json interface{}
	err = assets.GetBuildRequestSchema().ParseAndValidate(&json, b)
	if err != nil {
		return nil, err
	}
	switch json.(type) {
	case map[string]interface{}:
		break
	default:
		return nil, errors.New("build request must be a map")
	}
	jsonMap := json.(map[string]interface{})

	// prepare our request instance
	resources := make(map[string]*resourceImpl)
	request := requestImpl{
		id:            id,
		resources:     &resources,
		workspacePath: path.Join(workspacePath, id),
	}

	// build the resources map
	switch jsonMap["resources"].(type) {
	case map[string]interface{}:
		break
	default:
		return nil, errors.New("resources property must be a map of resource names to resource definitions")
	}
	for name, resourceJson := range jsonMap["resources"].(map[string]interface{}) {
		resourceJsonMap := resourceJson.(map[string]interface{})
		resources[name] = &resourceImpl{
			request:         &request,
			name:            name,
			resourceType:    resourceJsonMap["type"].(string),
			resourceConfig:  resourceJsonMap["config"],
			workspacePath:   path.Join(request.workspacePath, name),
			configSchema:    nil,
			initAction:      nil,
			discoveryAction: nil,
		}
		resources[name].initAction = &actionImpl{
			resource: resources[name],
			name:     "init",
			image:    resources[name].resourceType,
		}
	}
	return &request, nil
}
