package middleware

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo/v4"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	emfctx "github.com/cambridge-blockchain/emf/emf/context"
)

// DDMiddleware provides a middleware that performs tasks common to all endpoints.
type DDMiddleware struct {
	agentHost     string
	analyticsRate float64
	env           string
	serviceName   string
	stackName     string
	statsDClient  *statsd.Client

	analytics bool
	enabled   bool
}

// DDOption represents an option that can be passed to Middleware.
type DDOption func(*DDMiddleware)

// WithAgentHost specifies a custom agent host.
func WithAgentHost(agentHost string) DDOption {
	return func(dd *DDMiddleware) {
		if agentHost != "" {
			dd.agentHost = agentHost
		} else {
			dd.agentHost = "127.0.0.1:8126"
		}
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) DDOption {
	return func(dd *DDMiddleware) {
		if on {
			dd.analytics = true
			dd.analyticsRate = 1.0
		}
	}
}

// WithServiceName sets the given service name for the system.
func WithServiceName(name string) DDOption {
	return func(dd *DDMiddleware) {
		dd.serviceName = name
	}
}

// WithEnv sets the given env executing.
func WithEnv(env string) DDOption {
	return func(dd *DDMiddleware) {
		dd.env = env
	}
}

// WithStatsDMetrics specifies client to send custom metrics in middleware to agent.
func WithStatsDMetrics(enabled bool, statsdHost string, env string, app string) DDOption {
	return func(dd *DDMiddleware) {
		if enabled {
			if statsdHost == "" {
				statsdHost = "127.0.0.1:8125"
			}
			if env == "" {
				env = dd.env
			}
			if app == "" {
				app = dd.stackName
			}
			var err error
			// use general statsd client for getting aggregate counts, gauges & series across
			// components
			dd.statsDClient, err = statsd.New(env,
				statsd.WithNamespace(dd.stackName+"."), // prefix every metric with the app name
				statsd.WithTags([]string{env}),         // send zone here, etc
			)
			if err != nil {
				panic(err)
			}
		}
	}
}

// NewDDTracerMiddleware is a variadic constructor for a DDTrace.
func NewDDTracerMiddleware(enabled bool, opts ...DDOption) *DDMiddleware {
	if enabled {
		var dd = &DDMiddleware{
			enabled:   enabled,
			stackName: "idbridge",
		}
		for _, opt := range opts {
			opt(dd)
		}
		if dd.env == "" {
			dd.env = "dev-local"
		}
		k8sDDAgentHost := os.Getenv("DD_AGENT_HOST")
		if k8sDDAgentHost != "" {
			dd.agentHost = k8sDDAgentHost
		} else if dd.agentHost == "" {
			dd.agentHost = "127.0.0.1:8126"
		}
		tracer.Start(
			tracer.WithAgentAddr(dd.agentHost),
			tracer.WithAnalytics(dd.analytics),
			tracer.WithPrioritySampling(),
		)
		return dd
	}

	return nil
}

// Middleware returns echo middleware which will trace incoming requests.
func (dd *DDMiddleware) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var ectx = c.(emfctx.EMFContext)
		if dd.enabled {
			request := ectx.Request()
			if !strings.HasSuffix(c.Path(), infoPath) && !strings.HasSuffix(c.Path(), metricsPath) {
				resource := request.Method + " /" + dd.serviceName + c.Path()

				// handle special cases for use of X-Query-Id
				if strings.HasSuffix(c.Path(), "/update") || strings.HasSuffix(c.Path(), "/queries") {
					if queryID, exists := request.URL.Query()["queryId"]; exists && len(queryID) > 0 {
						resource = resource + " " + queryID[0]
					}
				}

				opts := []ddtrace.StartSpanOption{
					tracer.ServiceName(dd.serviceName),
					tracer.ResourceName(resource),
					tracer.SpanType(ext.SpanTypeWeb),
					tracer.Tag(ext.HTTPMethod, request.Method),
					tracer.Tag(ext.HTTPURL, request.URL.Path),
					tracer.Tag("stack", dd.stackName),
					tracer.Tag("requestId", ectx.GetRequestID()),
					tracer.Tag("env", dd.env),
					tracer.Tag(ext.AnalyticsEvent, true),
				}

				var spanctx ddtrace.SpanContext
				if spanctx, _ = tracer.Extract(tracer.HTTPHeadersCarrier(request.Header)); err == nil { //nolint:errcheck
					opts = append(opts, tracer.ChildOf(spanctx))
				}

				var span ddtrace.Span
				var ctx context.Context
				span, ctx = tracer.StartSpanFromContext(request.Context(), "http.request", opts...)
				defer span.Finish()
				// Investigate how to better use tracer.Inject ability in our stack to pushdown vars
				_ = tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(request.Header)) //nolint:errcheck
				ectx.Header().Set("X-Datadog-Parent-Id", request.Header.Get("X-Datadog-Parent-Id"))
				ectx.Header().Set("X-Datadog-Trace-Id", request.Header.Get("X-Datadog-Trace-Id"))
				// https://docs.datadoghq.com/tracing/guide/trace_sampling_and_storage/
				// -1	User input	The Agent drops the trace.
				// 0	Automatic sampling decision	The Agent drops the trace.
				// 1	Automatic sampling decision	The Agent keeps the trace.
				// 2	User input	The Agent keeps the trace, and the backend will only apply sampling if above maximum volume allowed.
				// https://godoc.org/gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer#Span.SetSamplingPriority
				ectx.Header().Set("X-Datadog-Sampling-Priority", "1")
				// pass the span through the request context
				ectx.SetRequest(request.WithContext(ctx))

				err = next(ectx)
				if err != nil {
					span.SetTag(ext.Error, err)
				}

				span.SetTag(ext.HTTPCode, strconv.Itoa(c.Response().Status))

				return err
			}
		}

		return next(c)
	}
}
