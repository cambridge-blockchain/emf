package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/emf/context"
	"github.com/cambridge-blockchain/emf/emf/logger"
)

// LoggingMiddleware provides a middleware that performs tasks common to all endpoints.
type LoggingMiddleware struct {
	logger echo.MiddlewareFunc
}

// LoggingOption provides the client a callback that is used to dynamically specify attributes for a
// LoggingMiddleware.
type LoggingOption func(*LoggingMiddleware)

// ElasticSearchOption takes in the ElasticSearch config boolean, and sets the LoggingMiddleware accordingly
func ElasticSearchOption(isElasticSearch bool) LoggingOption {
	return func(lm *LoggingMiddleware) {
		if isElasticSearch {
			lm.logger = logger.Middleware()
		} else {
			lm.logger = logger.DefaultLoggingMiddleware
		}
	}
}

// NewLoggingMiddleware is a variadic constructor for a LoggingMiddleware.
func NewLoggingMiddleware(opts ...LoggingOption) *LoggingMiddleware {
	var lm = &LoggingMiddleware{
		logger: logger.DefaultLoggingMiddleware,
	}

	for _, opt := range opts {
		opt(lm)
	}

	return lm
}

// Wrapper is a pass through function for handlers that implicitly performs additional business
// logic per request.
func (lm *LoggingMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var ctx = c.(context.EMFContext)
		var id = ctx.GetRequestID()

		ctx.Request().Header.Set(echo.HeaderXRequestID, id)
		ctx.Header().Set(echo.HeaderXRequestID, id)

		if ctx.Path() != infoPath && ctx.Path() != metricsPath {
			ctx.Logger().Debugf("request_id=%s | %s %s", id, ctx.Request().Method, ctx.Path())
		}

		return lm.logger(next)(ctx)
	}
}
