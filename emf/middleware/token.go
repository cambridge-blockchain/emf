package middleware

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/emf/context"
)

// TokenMiddleware provides a middleware that verifies required authentication for endpoints.
type TokenMiddleware struct {
	verify func(*jwt.Token) bool
}

// TokenOption provides the client a callback that is used to dynamically specify attributes for a
// TokenMiddleware.
type TokenOption func(*TokenMiddleware)

// WithTokenVerifier is used for specifying the verify function for JWT Tokens.
func WithTokenVerifier(verify func(*jwt.Token) bool) TokenOption {
	return func(tm *TokenMiddleware) { tm.verify = verify }
}

// NewTokenMiddleware is a variadic constructor for a TokenMiddleware.
func NewTokenMiddleware(opts ...TokenOption) *TokenMiddleware {
	tm := &TokenMiddleware{
		verify: func(*jwt.Token) bool {
			return true
		},
	}

	for _, opt := range opts {
		opt(tm)
	}

	return tm
}

// Wrapper is a pass through function for handlers that implicitly performs additional business
// logic per request.
func (tokenMid *TokenMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(c echo.Context) (err error) {
		var (
			token *jwt.Token
			ok    bool
		)

		ctx := c.(context.EMFContext)

		if token, ok = ctx.Get("user").(*jwt.Token); !ok || token == nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"JWT Token not set in Context, run Auth middleware first",
			)
		}

		if ok = tokenMid.verify(token); !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token for this request")
		}

		return next(ctx)
	})
}
