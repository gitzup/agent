package pkg

type HasWorkspace interface {
	GetWorkspacePath() string
}

type BuildRequest interface {
	HasWorkspace
	GetID() string
	GetResources() *map[string]Resource
}

type Resource interface {
	HasWorkspace
	GetBuildRequest() BuildRequest
	GetName() string
	GetType() string
	GetConfig() *interface{}
	Initialize() error
	DiscoverState() error
	Apply() error
}
