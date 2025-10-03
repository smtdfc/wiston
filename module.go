package wiston

type Module struct {
	Name string
}

// NewModule creates and returns a new Module with the given name.
func NewModule(name string) *Module {
	return &Module{
		Name: name,
	}
}
