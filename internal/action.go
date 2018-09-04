package internal

// Single action provided by a resource.
type Action struct {
	Resource   Resource
	Name       string
	Image      string
	Entrypoint []string
	Cmd        []string
}

func (action *Action) Invoke() error {
	// TODO: execute action
	return nil
}
