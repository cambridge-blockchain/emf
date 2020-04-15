package middleware

import (
	"crypto/rsa"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/cambridge-blockchain/emf/configurer"
	"github.com/cambridge-blockchain/emf/emf/context"
)

// AuthMiddleware provides a middleware that verifies required authentication for endpoints.
type AuthMiddleware struct {
	signingKey *rsa.PublicKey
	claims     jwt.Claims
	method     string
	component  string
	path       string
	payload    interface{}
	skipper    func(c echo.Context) bool
}

// AuthOption provides the client a callback that is used to dynamically specify attributes for an
// AuthMiddleware.
type AuthOption func(*AuthMiddleware)

// WithAuthClaims is used for specifying the Logger for an AuthMiddleware.
func WithAuthClaims(claims jwt.Claims) AuthOption {
	return func(am *AuthMiddleware) { am.claims = claims }
}

// WithActiveTokenRequest configures the request to an external component for token validation
func WithActiveTokenRequest(method string, component string, path string, payload interface{}) AuthOption {
	return func(am *AuthMiddleware) {
		am.method = method
		am.component = component
		am.path = path
		am.payload = payload
	}
}

// WithAuthSkipper configures the AuthMiddleware Skipper function.
// The Skipper function determines which endpoints skip authentication
func WithAuthSkipper(skipper func(c context.EMFContext) bool) AuthOption {
	return func(am *AuthMiddleware) {
		am.skipper = func(c echo.Context) bool {
			return skipper(c.(context.EMFContext))
		}
	}
}

// NewAuthMiddleware is a variadic constructor for an AuthMiddleware.
func NewAuthMiddleware(cfg configurer.ConfigReader, opts ...AuthOption) *AuthMiddleware {
	var (
		am   *AuthMiddleware
		pkey *rsa.PublicKey
		err  error
	)

	var publicKey = []byte(cfg.GetString("api.public_key"))

	if pkey, err = jwt.ParseRSAPublicKeyFromPEM(publicKey); err != nil {
		log.Fatal(err)
	}

	am = &AuthMiddleware{
		signingKey: pkey,
		claims:     jwt.MapClaims{},
		// Skip info and metrics endpoints by default
		skipper: func(c echo.Context) bool {
			switch c.Path() {
			case infoPath:
				fallthrough
			case metricsPath:
				return true
			default:
				if strings.HasPrefix(c.Path(), "/debug/pprof") {
					return true
				}
				if strings.HasPrefix(c.Path(), "/noauth") {
					return true
				}
				return false
			}
		},
	}

	for _, opt := range opts {
		opt(am)
	}

	return am
}

// Wrapper is a pass through function for handlers that implicitly performs additional business
// logic per request. In this case, we're wrapping an existing echo middleware with the right config.
func (am *AuthMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	conf := emiddleware.JWTConfig{
		Claims:        am.claims,
		SigningKey:    am.signingKey,
		SigningMethod: "RS256",
		Skipper:       am.skipper,
	}
	return emiddleware.JWTWithConfig(conf)(
		func(c echo.Context) (err error) {
			type authResponse struct {
				Active bool `json:"active"`
			}
			var resPayload authResponse

			var ctx = c.(context.EMFContext)

			if am.skipper(c) {
				return next(ctx)
			}

			var token *jwt.Token
			var ok bool
			if token, ok = ctx.Get("user").(*jwt.Token); !ok || token == nil {
				return ctx.NewError("emf.401.TokenVerificationFailure", map[string]interface{}{
					"error": err,
				})
			}
			ctx.Header().Add("Authorization", "Bearer "+token.Raw)

			if am.component == "" {
				ctx.Logger().Warn("No Active Token Request configured, skipping further JWT token checks...")
				return next(ctx)
			}

			if err = ctx.Requester(
				am.method,
				am.component,
				am.path,
				am.payload,
				&resPayload,
			); err != nil {
				return ctx.NewError("emf.401.TokenVerificationFailure", map[string]interface{}{
					"error": err,
				})
			}

			if !resPayload.Active {
				return ctx.NewError("emf.401.TokenInactive", nil)
			}
			return next(ctx)
		})
}
