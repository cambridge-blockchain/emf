package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/emf/context"
)

// Middleware is used to for providing additional logic on endpoints registered with a Controller.
type Middleware interface {
	Wrapper(next echo.HandlerFunc) echo.HandlerFunc
}

// MiddlewareFunc is equivalent to the echo MiddlewareFunc interface
// to avoid the middleware package having to import it
type MiddlewareFunc = echo.MiddlewareFunc

// HandlerFunc is equivalent to the echo HandlerFunc interface
// to avoid the middleware package having to import it
type HandlerFunc = func(ctx context.EMFContext) error

// EchoRouter sets up the Request Handlers for the server.
type EchoRouter interface {
	Add(string, string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route
	Any(string, echo.HandlerFunc, ...echo.MiddlewareFunc) []*echo.Route
	Pre(middleware ...echo.MiddlewareFunc)
	Use(middleware ...echo.MiddlewareFunc)
	// Group(string, ...echo.MiddlewareFunc) *echo.Group
}

// Router is a generic router type that is used to register and dispatch routes
type Router struct {
	echo   EchoRouter
	Logger echo.Logger
}

// Option provides the client a callback that is used to dynamically specify attributes for a
// Controller.
type Option func(*Router)

// HTTPClient issues HTTP Requests and performs error handling.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// WithRouter is an Option for specifying the Router for a Controller.
func WithRouter(router *echo.Echo) Option {
	return func(r *Router) {
		r.echo = router
		r.Logger = router.Logger
	}
}

// New is a variadic constructor for a Controller.
func New(opts ...Option) *Router {
	e := echo.New()
	r := &Router{
		echo:   e,
		Logger: e.Logger,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// NewGroup creates a new router group with prefix and optional group-level middleware.
func (r *Router) NewGroup(prefix string, m ...Middleware) (g *Group) {
	g = &Group{prefix: prefix, router: r}
	g.Use(m...)
	return
}

// UseGlobalMiddlewares registers global middlewares for all endpoints
func (r *Router) UseGlobalMiddlewares(m ...Middleware) {
	r.Use(MiddlewaresWrapper(m)...)
}

// Use exposes the echo.Use method to callers for registering middleware
func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.echo.Use(middleware...)
}

// Pre exposes the echo.Pre method to callers for registering pre-middlewares
func (r *Router) Pre(middleware ...MiddlewareFunc) {
	r.echo.Pre(middleware...)
}
