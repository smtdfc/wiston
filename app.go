package wison

type App struct {
	Modules map[string]*Module
	Gateway map[GatewayType]Gateway
}

func (a *App) AddModule(module *Module) {
	a.Modules[module.Name] = module
}

func (a *App) UseGateway(gateway Gateway) {
	gt := gateway.GetType()
	a.Gateway[gt] = gateway
	gateway.SetApp(a)
}

func (a *App) Start() error {
	return nil
}

func NewApp() *App {
	return &App{
		Modules: make(map[string]*Module),
	}
}
