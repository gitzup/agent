package pkg

import (
	"fmt"
	"log"
)

type Action struct {
	resource   Resource
	Name       string   `json:"name"`
	Image      string   `json:"image"`
	Entrypoint []string `json:"entrypoint"`
	Cmd        []string `json:"cmd"`
}

type Resource struct {
	request         *BuildRequest
	Name            string      `json:"name"`
	Type            string      `json:"type"`
	Config          interface{} `json:"config"`
	workspacePath   string
	configSchema    *Schema
	discoveryAction *Action
}

type ResourceInitRequest struct {
	Id       string    `json:"id"`
	Resource *Resource `json:"resource"`
}

type ResourceInitResponse struct {
	ConfigSchema    interface{} `json:"configSchema"`
	DiscoveryAction Action      `json:"discoveryAction"`
}

func (resource *Resource) Initialize() error {
	log.Printf("Initializing resource '%s'...\n", resource.Name)
	var response ResourceInitResponse
	initRequest := DockerRequest{
		Input: ResourceInitRequest{
			Id:       resource.request.Id,
			Resource: resource,
		},
		ContainerName: fmt.Sprintf("%s-%s-init", resource.request.Id, resource.Name),
		Env: []string{
			fmt.Sprintf("GITZUP=%t", true),
			fmt.Sprintf("RESOURCE_NAME=%s", resource.Name),
			fmt.Sprintf("RESOURCE_TYPE=%s", resource.Type),
		},
		Volumes:         map[string]struct{}{},
		Image:           resource.Type,
		PostExitHandler: CreateJsonResultParser("/gitzup/result.json", GetInitResponseSchema(), &response),
	}
	if err := initRequest.Execute(); err != nil {
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

func (action *Action) Invoke() error {
	// TODO: execute action
	return nil
}
