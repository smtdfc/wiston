package wison

type Module struct {
	Name string
}

func NewModule(name string) *Module {
	return &Module{
		Name: name,
	}
}
