package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/cambridge-blockchain/emf/configurer"
)

const (
	infoPath    = "/info"
	metricsPath = "/metrics"
	// UUIDRegex represents a UUID regular expression
	UUIDRegex = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	seconds10 = 10 * time.Second
	seconds30 = 30 * time.Second
)

// AllMiddlewares is a struct with one configured Middleware of each type
type AllMiddlewares struct {
	Auth            *AuthMiddleware
	BodyLimitConfig emiddleware.BodyLimitConfig
	Context         *ContextMiddleware
	DDTracer        *DDMiddleware
	Logging         *LoggingMiddleware
	ParamChecker    *ParamCheckerMiddleware
	RateLimit       *RateLimitMiddleware
	Token           *TokenMiddleware
	External        []echo.MiddlewareFunc
}

type middlewareUser interface {
	Pre(middleware ...echo.MiddlewareFunc)
	Use(middleware ...echo.MiddlewareFunc)
}

// UseMiddlewares registers all of the middlewares for use
func (am *AllMiddlewares) UseMiddlewares(e middlewareUser) {
	// ***********************************************
	// * Register the Middlewares for Use
	// ***********************************************
	e.Pre(emiddleware.RemoveTrailingSlash()) // Remove the Trailing / from a request before matching the path
	e.Use(am.Context.Wrapper)                // MUST register the Context middleware first
	e.Use(am.External...)                    // Register Recover middleware in case any in-house middlewares panic
	e.Use(emiddleware.BodyLimitWithConfig(am.BodyLimitConfig))
	e.Use(am.Logging.Wrapper, am.Auth.Wrapper) // MUST Register these middlewares after External and Context.
	// The Logging Middleware uses the header set by the echo RequestID() middleware to print IDs for each request
	if am.DDTracer != nil {
		e.Use(am.DDTracer.Middleware)
	}
}

// InitMiddlewares configures default middlewares, and returns them all as a struct for later configuration
func InitMiddlewares(conf configurer.ConfigReader) (am *AllMiddlewares) {
	var bodyLimit string
	if bodyLimit = conf.GetString("api.max_http_body"); bodyLimit == "" {
		bodyLimit = "10M"
	}

	// ***********************************************
	// * Expose Middlewares
	// ***********************************************
	am = &AllMiddlewares{
		Auth: NewAuthMiddleware(conf),
		BodyLimitConfig: emiddleware.BodyLimitConfig{
			Limit: bodyLimit,
		},
		// Set according to https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
		Context: NewContextMiddleware(conf, WithContextClient(
			&http.Client{
				Timeout: seconds10,
				Transport: &http.Transport{
					Dial: (&net.Dialer{
						Timeout:   seconds30,
						KeepAlive: seconds30,
					}).Dial,
					TLSHandshakeTimeout:   seconds10,
					ResponseHeaderTimeout: seconds10,
					ExpectContinueTimeout: time.Second,
				},
			},
		)),
		DDTracer: NewDDTracerMiddleware(conf.GetBool("tracing.datadog"),
			WithEnv(conf.GetString("tracing.env")),
			WithServiceName(conf.GetString("api.service")),
			WithAnalytics(conf.GetBool("tracing.datadog_analytics")),
			WithAgentHost(conf.GetString("tracing.datadog_agent")),
		),
		Logging: NewLoggingMiddleware(
			ElasticSearchOption(conf.GetBool("logging.elasticsearch")),
		),
		ParamChecker: NewParamCheckerMiddleware(WithRegex(UUIDRegex + "|^[0-9]+$")),
		RateLimit:    NewRateLimitMiddleware(),
		Token:        NewTokenMiddleware(),
		External: []echo.MiddlewareFunc{
			emiddleware.Recover(),
			emiddleware.RequestID(),
		},
	}
	return
}
