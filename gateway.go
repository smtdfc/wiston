package wiston

import (
	"net/http"
	"sync"
)

// GatewayType defines the type of a gateway, such as HTTP or WebSocket.
type GatewayType int

const (
	// HTTP_GATEWAY represents an HTTP server gateway.
	HTTP_GATEWAY GatewayType = iota
	// WS_GATEWAY represents a WebSocket server gateway.
	WS_GATEWAY
)

// Gateway is the fundamental interface for network gateways.
// It defines the basic methods that all gateway types must implement.
type Gateway interface {
	// GetType returns the specific type of the gateway.
	GetType() GatewayType
	// SetApp associates the main application instance with the gateway.
	SetApp(app *App) error
	// Start launches the gateway, making it ready to accept connections.
	Start() error
}

// HttpGateway defines the interface for an HTTP gateway, extending the base Gateway.
// It provides methods for middleware and creating routing scopes.
type HttpGateway interface {
	Gateway
	// Use applies one or more middleware handlers to the gateway.
	Use(mw ...HttpHandler)
	// CreateScope creates a new routing scope for a specific module under a given prefix.
	CreateScope(module *Module, prefix string) HttpScope
}

// HttpStatus provides a convenient struct for accessing standard HTTP status codes.
var HttpStatus = struct {
	OK                  int
	Created             int
	Accepted            int
	NoContent           int
	BadRequest          int
	Unauthorized        int
	Forbidden           int
	NotFound            int
	MethodNotAllowed    int
	Conflict            int
	InternalServerError int
	NotImplemented      int
	BadGateway          int
	ServiceUnavailable  int
	GatewayTimeout      int
}{
	OK:                  http.StatusOK,
	Created:             http.StatusCreated,
	Accepted:            http.StatusAccepted,
	NoContent:           http.StatusNoContent,
	BadRequest:          http.StatusBadRequest,
	Unauthorized:        http.StatusUnauthorized,
	Forbidden:           http.StatusForbidden,
	NotFound:            http.StatusNotFound,
	MethodNotAllowed:    http.StatusMethodNotAllowed,
	Conflict:            http.StatusConflict,
	InternalServerError: http.StatusInternalServerError,
	NotImplemented:      http.StatusNotImplemented,
	BadGateway:          http.StatusBadGateway,
	ServiceUnavailable:  http.StatusServiceUnavailable,
	GatewayTimeout:      http.StatusGatewayTimeout,
}

// HttpContext defines the interface for the context of an HTTP request.
// It provides methods to access request data and to write a response.
type HttpContext interface {
	// Request methods
	Method() string
	Path() string
	Protocol() string

	Param(key string) string
	Query(key string) string
	QueryDefault(key, def string) string
	Header(key string) string
	Cookie(name string) string
	Body() []byte
	FormValue(key string) string
	FormFile(name string) ([]byte, error)

	// Response methods
	Status(code int) HttpContext
	SetHeader(key, value string) HttpContext
	SetCookie(name, value string, options ...any) HttpContext

	Text(code int, data string) HttpContext
	JSON(code int, data any) HttpContext
	HTML(code int, html string) HttpContext
	Blob(code int, contentType string, data []byte) HttpContext
	File(code int, filepath string) HttpContext

	// Flow control methods
	Next() HttpContext
	Abort() HttpContext
	IsAborted() bool

	// Data sharing methods
	Set(key string, value any) HttpContext
	Get(key string) any
	MustGet(key string) any
}

// HttpHandler defines the function signature for handling HTTP requests.
type HttpHandler func(HttpContext)

// HttpScope provides an interface for defining a group of routes
// under a common path prefix and with shared middleware.
type HttpScope interface {
	Use(mw ...HttpHandler)
	SetLogger(logger *Logger)
	Get(path string, handlers ...HttpHandler)
	Post(path string, handlers ...HttpHandler)
	Put(path string, handlers ...HttpHandler)
	Delete(path string, handlers ...HttpHandler)
	Patch(path string, handlers ...HttpHandler)
	Head(path string, handlers ...HttpHandler)
	Options(path string, handlers ...HttpHandler)
	Connect(path string, handlers ...HttpHandler)
	Trace(path string, handlers ...HttpHandler)
}

// WsContext defines the interface for the context of a WebSocket event.
// It provides methods for interacting with clients, rooms, and event data.
type WsContext interface {
	GetData() any
	GetEvent() string
	GetClientID() string
	Join(name string) error
	Leave(name string) error
	HasRoom(name string) bool
	HasJoin(name string) bool
	GetAllRoom() []string
	CreateRoom(name string) error
	EmitToRoom(room string, event string, data any) error
	Emit(event string, data any) error
}

// WsHandler defines the function signature for handling WebSocket events.
type WsHandler func(WsContext)

// WsNamespace provides an interface for defining a logical grouping
// of WebSocket event handlers.
type WsNamespace interface {
	SetLogger(logger *Logger)
	On(event string, handlers ...WsHandler)
}

// WsGateway defines the interface for a WebSocket gateway, extending the base Gateway.
// It provides methods for broadcasting messages and managing rooms and namespaces.
type WsGateway interface {
	Gateway
	GetAllRoom() []string
	Broadcast(event string, data any)
	CreateRoom(name string) error
	HasRoom(name string) bool
	CreateNamespace(module *Module, name string) WsNamespace
}

// GatewayManager manages the lifecycle of all registered gateways within an application.
type GatewayManager struct {
	App     *App
	Gateway map[GatewayType]Gateway
}

// UseGateway registers a new Gateway with the manager and associates it with the app.
func (g *GatewayManager) UseGateway(gateway Gateway) {
	gt := gateway.GetType()
	g.Gateway[gt] = gateway
	gateway.SetApp(g.App)
}

// StartAll launches all registered gateways concurrently.
// It returns a sync.WaitGroup that callers can use to wait for all gateways to stop.
func (g *GatewayManager) StartAll() *sync.WaitGroup {
	var wg sync.WaitGroup
	if g.Gateway[HTTP_GATEWAY] != nil {
		httpGateway := g.Gateway[HTTP_GATEWAY]
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.App.Logger.Info("Starting HTTP Gateway...")
			if err := httpGateway.Start(); err != nil {
				g.App.Logger.Error("HTTP Gateway failed: " + err.Error())
			}
			g.App.Logger.Info("HTTP Gateway stopped")
		}()
	}

	if g.Gateway[WS_GATEWAY] != nil {
		wsGateway := g.Gateway[WS_GATEWAY]
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.App.Logger.Info("Starting WS Gateway...")
			if err := wsGateway.Start(); err != nil {
				g.App.Logger.Error("WS Gateway failed: " + err.Error())
			}
			g.App.Logger.Info("WS Gateway stopped")
		}()
	}

	return &wg
}
