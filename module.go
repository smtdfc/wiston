package wiston

// Module represents a unit of functionality that can be
// registered or composed within the application framework.
type Module struct {
	// Name is the identifier of the module.
	Name string
}

// NewModule creates a new Module with the given name.
func NewModule(name string) *Module {
	return &Module{
		Name: name,
	}
}
