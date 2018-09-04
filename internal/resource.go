package internal

import (
	"fmt"
	"log"
	"path"

	"github.com/gitzup/agent/pkg"
)

// Resource to be applied in a build request.
type Resource struct {
	request         *BuildRequest
	Name            string       `json:"name"`
	Type            string       `json:"type"`
	Config          *interface{} `json:"config"`
	configSchema    *Schema
	discoveryAction *Action
}

// Request for resource initialization. This struct sent to a resource on its initialization request via resource.Initialize()
type ResourceInitRequest struct {
	RequestId string                      `json:"requestId"`
	Resource  ResourceInitRequestResource `json:"resource"`
}

// Resource information in a resource initialization request.
type ResourceInitRequestResource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Response from resource initialization. This struct is sent back from the resource.
type ResourceInitResponse struct {
	ConfigSchema    interface{} `json:"configSchema"`
	DiscoveryAction Action      `json:"discoveryAction"`
}

// Returns the name of the resource.
func (resource *Resource) GetName() string {
	return resource.Name
}

// Returns the type of the resource. Resource types are always references to Docker images.
func (resource *Resource) GetType() string {
	return resource.Type
}

// Returns the resource configuration, as it was sent in the build request for this resource.
func (resource *Resource) GetConfig() *interface{} {
	return resource.Config
}

// Returns the path of the build workspace for the resource.
func (resource *Resource) GetWorkspacePath() string {
	return path.Join(resource.request.GetWorkspacePath(), "resources", resource.GetName())
}

// Returns the build request this resource belongs to.
func (resource *Resource) GetBuildRequest() pkg.BuildRequest {
	return resource.request
}

func (resource *Resource) Initialize() error {
	log.Printf("Initializing resource '%s'...\n", resource.Name)
	var response ResourceInitResponse
	initRequest := DockerRequest{
		Input: ResourceInitRequest{
			RequestId: resource.request.Id,
			Resource: ResourceInitRequestResource{
				Name: resource.Name,
				Type: resource.Type,
			},
		},
		ContainerName: fmt.Sprintf("%s-%s-init", resource.request.Id, resource.Name),
		Env: []string{
			fmt.Sprintf("GITZUP=%t", true),
			fmt.Sprintf("GITZUP_VERSION=%s", Version),
			fmt.Sprintf("RESOURCE_NAME=%s", resource.Name),
			fmt.Sprintf("RESOURCE_TYPE=%s", resource.Type),
		},
		Volumes:         map[string]struct{}{},
		Image:           resource.Type,
		PostExitHandler: CreateJsonResultParser("/gitzup/result.json", GetInitResponseSchema(), &response),
	}
	if err := initRequest.InvokeAndDoWithContainer(); err != nil {
		return err
	}
	resourceConfigSchema, err := NewSchema(
		response.ConfigSchema,
		GetActionSchema(),
		GetResourceSchema(),
		GetBuildRequestSchema(),
		GetBuildResponseSchema(),
		GetInitRequestSchema(),
		GetInitResponseSchema(),
	)
	if err != nil {
		return err
	}
	resource.configSchema = resourceConfigSchema
	resource.discoveryAction = &response.DiscoveryAction

	if err := resource.configSchema.Validate(resource.Config); err != nil {
		return err
	}

	return nil
}

func (resource *Resource) DiscoverState() error {
	log.Printf("Discovering state for resource '%s'...\n", resource.Name)
	// TODO: implement resource DiscoverState()
	return nil
}

func (resource *Resource) Apply() error {
	// TODO: implement resource.Apply()
	return nil
}
