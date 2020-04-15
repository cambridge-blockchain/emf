package router

import (
	"net/http"
	"path"

	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/emf/context"
)

type (
	// Group is a set of sub-routes for a specified route. It can be used for inner
	// routes that share a common middleware or functionality that should be separate
	// from the parent echo instance while still inheriting from it.
	Group struct {
		prefix     string
		middleware []Middleware
		router     *Router
	}
)

// HandlerWrapper wraps EMFHandlerFuncs into echo.HandlerFuncs
func HandlerWrapper(emfhandler HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return emfhandler(c.(context.EMFContext))
	}
}

// GoHandlerToEMFHandler converts a standard go http.Handler function into a router.HandlerFunc
func GoHandlerToEMFHandler(gohandler http.Handler) HandlerFunc {
	return func(c context.EMFContext) error {
		return echo.WrapHandler(gohandler)(c)
	}
}

// MiddlewaresWrapper takes in a list of Middlewares and returns a list of echo.MiddlewareFuncs
func MiddlewaresWrapper(mids []Middleware) []echo.MiddlewareFunc {
	var (
		mid      Middleware
		midFuncs []echo.MiddlewareFunc
	)

	for _, mid = range mids {
		midFuncs = append(midFuncs, mid.Wrapper)
	}
	return midFuncs
}

// Use implements `Echo#Use()` for sub-routes within the Group.
func (g *Group) Use(middleware ...Middleware) {
	g.middleware = append(g.middleware, middleware...)
	// Allow all requests to reach the group as they might get dropped if router
	// doesn't find a match, making none of the group middleware process.
	for _, p := range []string{"", "/*"} {
		g.router.echo.Any(path.Clean(g.prefix+p), echo.NotFoundHandler, MiddlewaresWrapper(g.middleware)...)
	}
}

// CONNECT implements `Echo#CONNECT()` for sub-routes within the Group.
func (g *Group) CONNECT(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.CONNECT, path, h, m...)
}

// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
func (g *Group) DELETE(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.DELETE, path, h, m...)
}

// GET implements `Echo#GET()` for sub-routes within the Group.
func (g *Group) GET(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.GET, path, h, m...)
}

// HEAD implements `Echo#HEAD()` for sub-routes within the Group.
func (g *Group) HEAD(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.HEAD, path, h, m...)
}

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
func (g *Group) OPTIONS(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.OPTIONS, path, h, m...)
}

// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
func (g *Group) PATCH(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.PATCH, path, h, m...)
}

// POST implements `Echo#POST()` for sub-routes within the Group.
func (g *Group) POST(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.POST, path, h, m...)
}

// PUT implements `Echo#PUT()` for sub-routes within the Group.
func (g *Group) PUT(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.PUT, path, h, m...)
}

// TRACE implements `Echo#TRACE()` for sub-routes within the Group.
func (g *Group) TRACE(path string, h HandlerFunc, m ...Middleware) *echo.Route {
	return g.Add(echo.TRACE, path, h, m...)
}

var (
	methods = [...]string{
		echo.CONNECT,
		echo.DELETE,
		echo.GET,
		echo.HEAD,
		echo.OPTIONS,
		echo.PATCH,
		echo.POST,
		echo.PROPFIND,
		echo.PUT,
		echo.TRACE,
	}
)

// Any implements `Echo#Any()` for sub-routes within the Group.
func (g *Group) Any(path string, h HandlerFunc, m ...Middleware) []*echo.Route {
	routes := make([]*echo.Route, len(methods))
	for i, method := range methods {
		routes[i] = g.Add(method, path, h, m...)
	}
	return routes
}

// Match implements `Echo#Match()` for sub-routes within the Group.
func (g *Group) Match(methods []string, path string, h HandlerFunc, m ...Middleware) []*echo.Route {
	routes := make([]*echo.Route, len(methods))
	for i, method := range methods {
		routes[i] = g.Add(method, path, h, m...)
	}
	return routes
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, middleware ...Middleware) *Group {
	m := make([]Middleware, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.router.NewGroup(g.prefix+prefix, m...)
}

// Add implements `Echo#Add()` for sub-routes within the Group.
func (g *Group) Add(method, path string, h HandlerFunc, middleware ...Middleware) *echo.Route {
	// Combine into a new slice to avoid accidentally passing the same slice for
	// multiple routes, which would lead to later add() calls overwriting the
	// middleware from earlier calls.
	m := make([]Middleware, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.router.echo.Add(method, g.prefix+path, HandlerWrapper(h), MiddlewaresWrapper(m)...)
}
