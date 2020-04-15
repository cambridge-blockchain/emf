package middleware

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/labstack/echo/v4"
)

// RateLimitMiddleware provides a middleware that verifies required authentication for endpoints.
type RateLimitMiddleware struct {
	limiter *limiter.Limiter
}

// RateLimitOption provides the client a callback that is used to dynamically specify attributes for a
// RateLimitMiddleware.
type RateLimitOption func(*RateLimitMiddleware)

// WithLimiter is used for specifying the rate to limit.
func WithLimiter(lmt *limiter.Limiter) RateLimitOption {
	return func(rlm *RateLimitMiddleware) { rlm.limiter = lmt }
}

// NewRateLimitMiddleware is a variadic constructor for a RateLimitMiddleware.
func NewRateLimitMiddleware(opts ...RateLimitOption) *RateLimitMiddleware {
	const oneRequestPerSecond = 1
	rlm := &RateLimitMiddleware{
		limiter: tollbooth.NewLimiter(oneRequestPerSecond, nil),
	}

	for _, opt := range opts {
		opt(rlm)
	}

	return rlm
}

// Wrapper is a pass through function for handlers that implicitly performs additional business
// logic per request.
func (rlm *RateLimitMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(c echo.Context) error {
		httpError := tollbooth.LimitByRequest(rlm.limiter, c.Response(), c.Request())
		if httpError != nil {
			return echo.NewHTTPError(httpError.StatusCode, httpError.Message)
		}
		return next(c)
	})
}
