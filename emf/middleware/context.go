package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/configurer"
	"github.com/cambridge-blockchain/emf/emf/context"
)

// ContextMiddleware provides a middleware that performs tasks common to all endpoints.
type ContextMiddleware struct {
	cfg  configurer.ConfigReader
	opts []context.Option
}

// ContextOption provides the client a callback that is used to dynamically specify attributes for a
// ContextMiddleware.
type ContextOption func(*ContextMiddleware)

// WithContextClient is used to specify the HTTP Client for the Requester to use.
func WithContextClient(client *http.Client) ContextOption {
	return func(cm *ContextMiddleware) {
		cm.opts = append(cm.opts, context.WithContextClient(client))
	}
}

// NewContextMiddleware is a variadic constructor for a ContextMiddleware.
func NewContextMiddleware(cfg configurer.ConfigReader, opts ...ContextOption) *ContextMiddleware {
	var cm = &ContextMiddleware{
		cfg: cfg,
	}

	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

// Wrapper is the middleware function itself which redefines the echo.Context to our Custom object
func (cm *ContextMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(context.NewEMFContext(
			c, cm.cfg, cm.opts...,
		))
	}
}
