package wiston

import (
	"net/http"
	"sync"
)

type GatewayType int

const (
	HTTP_GATEWAY GatewayType = iota
	WS_GATEWAY
)

type Gateway interface {
	GetType() GatewayType
	SetApp(app *App) error
	Start() error
}

type HttpGateway interface {
	Gateway
	Use(mw ...HttpHandler)
	CreateScope(module *Module, prefix string) HttpScope
}

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

type HttpContext interface {
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

	Status(code int) HttpContext
	SetHeader(key, value string) HttpContext
	SetCookie(name, value string, options ...any) HttpContext

	Text(code int, data string) HttpContext
	JSON(code int, data any) HttpContext
	HTML(code int, html string) HttpContext
	Blob(code int, contentType string, data []byte) HttpContext
	File(code int, filepath string) HttpContext

	Next() HttpContext
	Abort() HttpContext
	IsAborted() bool

	Set(key string, value any) HttpContext
	Get(key string) any
	MustGet(key string) any
}

type HttpHandler func(HttpContext)

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

type WsHandler func(WsContext)

type WsNamespace interface {
	SetLogger(logger *Logger)
	On(event string, handlers ...WsHandler)
}

type WsGateway interface {
	Gateway
	GetAllRoom() []string
	Broadcast(event string, data any)
	CreateRoom(name string) error
	HasRoom(name string) bool
	CreateNamespace(module *Module, name string) WsNamespace
}

type GatewayManager struct {
	App     *App
	Gateway map[GatewayType]Gateway
}

// UseGateway registers a new Gateway with the manager.
func (g *GatewayManager) UseGateway(gateway Gateway) {
	gt := gateway.GetType()
	g.Gateway[gt] = gateway
	gateway.SetApp(g.App)
}

// StartAll launches all registered gateways concurrently and returns a WaitGroup that can be used to wait for them to stop.
func (g *GatewayManager) StartAll() sync.WaitGroup {
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

	return wg
}
