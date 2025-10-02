package wison

type GatewayType int

const (
	HTTP_GATEWAY GatewayType = iota
	WS_GATEWAY
)

type Gateway interface {
	GetType() GatewayType
	SetApp(app *App) error
}

type HttpGateway interface {
	Gateway
}

func ResolveGateway[T Gateway](app *App, gatewayType GatewayType) T {
	gw, ok := app.Gateway[gatewayType].(T)
	if !ok {
		var zero T
		return zero
	}
	return gw
}
