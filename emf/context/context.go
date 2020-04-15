package context

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/configurer"
	"github.com/cambridge-blockchain/emf/emf/context/errors"
)

// EMFContext is the Interface which is passed into each EMF request handler as part of each request
type EMFContext interface {
	echo.Context
	LoggerlessRequestHandler
	GetClaim(claim string) (c string, err error)
	GetRequestID() string
	GetRequestHandler() RequestHandler
	GetLimitAndOffset() (limit int, offset int, err error)
	GetRequestLimited(domain, path string) (req *http.Request, err error)
}

// EMFContextType is the default implementation of the EMFContext interface
type EMFContextType struct {
	echo.Context
	RequestHandler
}

// Option provides the client a callback that is used to dynamically specify attributes for a
// EMFContext.
type Option func(*EMFContextType)

// WithContextClient is used to specify the HTTP Client for the Requester to use.
func WithContextClient(client *http.Client) Option {
	return func(ctx *EMFContextType) {
		ctx.RequestHandler.(*RequestHandlerType).client = client
	}
}

// WithRequestHandler is used to modify the default RequestHandler object
func WithRequestHandler(rh RequestHandler) Option {
	return func(ctx *EMFContextType) { ctx.RequestHandler = rh }
}

// NewEMFContext is a variadic constructor for a EMFContext.
func NewEMFContext(c echo.Context, cfg configurer.ConfigReader, opts ...Option) (ctx *EMFContextType) {
	rh := &RequestHandlerType{
		cfg,
		&http.Client{},
		nil,
		http.Header{},
	}
	ctx = &EMFContextType{
		c,
		rh,
	}
	rh.eh = NewEMFErrorHandler(ctx,
		ctx.IsDebug(),
		errors.WithLogger(ctx.Context.Logger()),
		errors.WithTemplate(cfg.GetString("errors.configPath")),
	)

	for _, opt := range opts {
		opt(ctx)
	}
	return
}

// NewEMFErrorHandler is a variadic constructor for a EMFErrorHandler.
func NewEMFErrorHandler(c echo.Context, isDebug bool, opts ...errors.HandlerOption) (eh *errors.EMFErrorHandlerType) {
	eh = &errors.EMFErrorHandlerType{
		DebugMode:   isDebug,
		Method:      c.Request().Method,
		Path:        c.Path(),
		QueryString: c.QueryString(),
	}
	for _, opt := range opts {
		opt(eh)
	}
	return
}

// Logger is wrapper to disambiguate the Logger method
func (ctx *EMFContextType) Logger() echo.Logger {
	return ctx.Context.Logger()
}

// ErrorHandler is a helper function to make an ErrorHandler object from the Context and return it
func (ctx *EMFContextType) ErrorHandler() errors.EMFErrorHandler {
	return ctx.RequestHandler.ErrorHandler()
}

// NewError is a helper function to wrap ErrorHandler.NewError
func (ctx *EMFContextType) NewError(code string, data map[string]interface{}, errors ...error) error {
	return ctx.ErrorHandler().NewError(code, data, errors...)
}

// GetRequestHandler is a getter for the RequestHandler object embedded in the context
func (ctx *EMFContextType) GetRequestHandler() RequestHandler {
	return ctx.RequestHandler
}

// IsDebug is a helper function to provide access to the Debug Mode config flag
func (ctx *EMFContextType) IsDebug() bool {
	return ctx.RequestHandler.IsDebug() || ctx.QueryParam("debug_mode") == "true"
}

// GetClaim is a helper function to handle type and map checking for JWT Claims
func (ctx *EMFContextType) GetClaim(claim string) (c string, err error) {
	var (
		token  *jwt.Token
		claims jwt.MapClaims
		ok     bool
		i      interface{}
	)

	if token = ctx.Get("user").(*jwt.Token); token == nil {
		err = fmt.Errorf("invalid JWT Token in Context, run Auth middleware first")
		return
	}

	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		err = fmt.Errorf("invalid JWT Token claims in Context, run Auth middleware first")
		return
	}
	if i, ok = claims[claim]; !ok {
		err = fmt.Errorf("JWT Claims do not include field: '%s'", claim)
		return
	}

	// Type Switch to support Int and Bool values for token claims
	switch val := i.(type) {
	case string:
		c = val
		return
	case int:
		c = strconv.Itoa(val)
		return
	case bool:
		c = strconv.FormatBool(val)
		return
	case float64:
		c = strconv.FormatFloat(val, 'f', 0, 64)
		return
	default:
		err = fmt.Errorf("JWT Claim '%s' is not of type string. Value: %v", claim, val)
		return
	}
}

// GetRequestID is a helper function to expose the RequestID Header
func (ctx *EMFContextType) GetRequestID() string {
	return ctx.Response().Header().Get(echo.HeaderXRequestID)
}

// GetRequestLimited is the method to set up a HTTP Get request with no body and a limit Query Parameter
func (ctx *EMFContextType) GetRequestLimited(domain, path string) (req *http.Request, err error) {
	var limit, offset int
	if limit, offset, err = ctx.GetLimitAndOffset(); err != nil {
		return
	}

	var u *url.URL
	if u, err = url.Parse(path); err != nil {
		err = ctx.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
			"Param": "-",
			"Error": fmt.Errorf("failed to parse path '%s'", path),
		})
		return
	}

	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	return ctx.GetRequestHandler().InitRequest(EncodingGob, http.MethodGet, domain, u.String(), nil)
}

// GetLimitAndOffset is a helper function to parse limit and offset query parameters and return them
func (ctx *EMFContextType) GetLimitAndOffset() (limit int, offset int, err error) {
	if limStr := ctx.QueryParam("limit"); limStr != "" {
		if limit, err = strconv.Atoi(limStr); err != nil {
			err = ctx.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
				"Param": "limit",
				"Error": fmt.Errorf("limit '%v' is not a valid integer", limit),
			})
			return
		}

		if limit < 0 || limit > ctx.GetMaxLimit() {
			err = ctx.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
				"Param": "limit",
				"Error": fmt.Errorf("limit '%v' must be between 1 and %v", limit, ctx.GetMaxLimit()),
			})
			return
		}
	}

	if limit == 0 {
		limit = ctx.GetDefaultLimit()
	}

	if offStr := ctx.QueryParam("offset"); offStr != "" {
		if offset, err = strconv.Atoi(offStr); err != nil {
			err = ctx.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
				"Param": "offset",
				"Error": fmt.Errorf("offset '%v' is not a valid integer", offset),
			})
			return
		}

		if offset < 0 {
			err = ctx.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
				"Param": "offset",
				"Error": fmt.Errorf("offset '%v' is not a positive integer", offset),
			})
			return
		}
	}
	return
}
