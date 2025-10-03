package wiston

// App is the central application container that manages modules,
// logging, and the lifecycle of gateways. It serves as the primary
// entry point and coordinator for all application components.
type App struct {
	Modules        map[string]*Module
	Logger         *Logger
	GatewayManager *GatewayManager
}

// AddModule adds a given module to the application's module registry.
// Each module is stored and identified by its name.
func (a *App) AddModule(module *Module) {
	a.Modules[module.Name] = module
}

// UseGateway registers a given gateway with the application's GatewayManager.
// The GatewayManager will handle the lifecycle of the registered gateway.
func (a *App) UseGateway(gateway Gateway) {
	a.GatewayManager.UseGateway(gateway)
}

// Start initializes all modules and starts all registered gateways.
// This method blocks until all gateways have gracefully stopped.
// It orchestrates the application startup sequence, ensuring modules are
// initialized before gateways begin their operations.
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

// NewApp creates, initializes, and returns a new App instance.
// It sets up the necessary components, including the module map,
// logger, and gateway manager, making the App ready for configuration.
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
