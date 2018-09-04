package internal

import (
	"log"
	"path"

	"github.com/gitzup/agent/pkg"
)

// Build request.
type BuildRequest struct {
	Id        string                `json:"id"`
	Resources *map[string]*Resource `json:"resources"`
}

// Returns the build request's ID.
func (request *BuildRequest) GetID() string {
	return request.Id
}

// Returns the build request's resources map, where each key is a resource's name, and the value is the resource instance.
func (request *BuildRequest) GetResources() *map[string]pkg.Resource {
	var resources = make(map[string]pkg.Resource)
	for name, resource := range *request.Resources {
		resources[name] = resource
	}
	return &resources
}

// Returns the build request's local workspace path.
func (request *BuildRequest) GetWorkspacePath() string {
	return path.Join(Config.Workspace, request.Id)
}

// Parse the build request from the given bytes array.
func ParseBuildRequest(id string, b []byte) (*BuildRequest, error) {
	var req BuildRequest

	// validate & parse the build request
	err := GetBuildRequestSchema().ParseAndValidate(&req, b)
	if err != nil {
		return nil, err
	}

	// post-parsing updates
	req.Id = id
	for name, resource := range *req.Resources {
		resource.request = &req
		resource.Name = name
	}
	return &req, nil
}

// Apply the build request
func (request *BuildRequest) Apply() error {
	log.Printf("Applying build request '%s'", request.Id)

	for _, resource := range *request.Resources {
		err := resource.Initialize()
		if err != nil {
			return err
		}
	}

	for _, resource := range *request.Resources {
		err := resource.DiscoverState()
		if err != nil {
			return err
		}
	}

	return nil
}
