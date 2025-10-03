package wiston

// App is the central application container that manages modules,
// logging, and gateway lifecycle.
type App struct {
	Modules        map[string]*Module
	Logger         *Logger
	GatewayManager *GatewayManager
}

// AddModule adds the given module to the application.
func (a *App) AddModule(module *Module) {
	a.Modules[module.Name] = module
}

// UseGateway registers the given gateway with the application.
func (a *App) UseGateway(gateway Gateway) {
	a.GatewayManager.UseGateway(gateway)
}

// Start initializes all modules and starts all registered gateways.
// It blocks until all gateways have stopped.
func (a *App) Start() error {
	a.Logger.Info("Starting application...")

	for name := range a.Modules {
		a.Logger.Info("Initializing module: " + name)
		// if len(module.onStartCallbacks) > 0 {
		// 	module.triggerHook("start")
		// }
		a.Logger.Info("Module " + name + " initialized")
	}

	// Start all gateways
	wg := a.GatewayManager.StartAll()

	// Wait until all gateways exit
	wg.Wait()
	return nil
}

// NewApp creates and returns a new App.
func NewApp() *App {
	app := &App{
		Modules: make(map[string]*Module),
		Logger:  &Logger{},
	}

	app.GatewayManager = &GatewayManager{
		App:     app,
		Gateway: make(map[GatewayType]Gateway),
	}

	return app
}
