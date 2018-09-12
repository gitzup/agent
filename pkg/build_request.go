package pkg

import (
	"log"
	"path"
)

type BuildRequest struct {
	Id            string                `json:"id"`
	Resources     *map[string]*Resource `json:"resources"`
	workspacePath string
}

func ParseBuildRequest(id string, b []byte, workspacePath string) (*BuildRequest, error) {
	var req BuildRequest

	// validate & parse the build request
	err := GetBuildRequestSchema().ParseAndValidate(&req, b)
	if err != nil {
		return nil, err
	}

	// post-parsing updates
	req.Id = id
	req.workspacePath = path.Join(workspacePath, req.Id)
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
