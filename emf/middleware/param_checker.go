package middleware

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/emf/context"
)

// ParamCheckerMiddleware is used for checking that the parameters are valid UUIDs or integers.
type ParamCheckerMiddleware struct {
	regex *regexp.Regexp
}

// ParamOption provides the client a callback that is used to dynamically specify attributes for a
// ParamCheckerMiddleware.
type ParamOption func(*ParamCheckerMiddleware)

// WithRegex is used for specifying the regex to check params against.
func WithRegex(regex string) ParamOption {
	return func(pcm *ParamCheckerMiddleware) {
		var err error
		if pcm.regex, err = regexp.Compile(regex); err != nil {
			panic(err)
		}
	}
}

// NewParamCheckerMiddleware is a variadic constructor for a TokenMiddleware.
func NewParamCheckerMiddleware(opts ...ParamOption) *ParamCheckerMiddleware {
	pcm := &ParamCheckerMiddleware{}

	for _, opt := range opts {
		opt(pcm)
	}

	return pcm
}

// Wrapper is middleware for checking that the parameters are valid UUIDs or integers.
func (pcm ParamCheckerMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		for _, param := range c.ParamValues() {
			if strings.ContainsRune(param, '/') {
				return echo.ErrNotFound
			}
			if match := pcm.regex.MatchString(param); !match {
				ctx := c.(context.EMFContext)
				return ctx.NewError("emf.400.InvalidParametersFailure", map[string]interface{}{
					"Error": fmt.Errorf("parameter '%s' is not a valid UUID or integer", param),
				})
			}
		}

		return next(c)
	}
}
