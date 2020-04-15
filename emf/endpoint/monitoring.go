package endpoint

import (
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cambridge-blockchain/emf/emf/router"
	"github.com/cambridge-blockchain/emf/models"
)

// RegisterMonitoring registers the prometheus monitoring endpoints
func RegisterMonitoring(r *models.Router, mids ...models.Middleware) {
	var g = r.NewGroup("/debug/pprof", mids...)
	g.Any("", router.GoHandlerToEMFHandler(pprof.Handler("index")))
	g.Any("/cmdline", router.GoHandlerToEMFHandler(pprof.Handler("cmdline")))
	g.Any("/profile", router.GoHandlerToEMFHandler(pprof.Handler("profile")))
	g.Any("/symbol", router.GoHandlerToEMFHandler(pprof.Handler("symbol")))
	g.Any("/trace", router.GoHandlerToEMFHandler(pprof.Handler("trace")))

	var g2 = r.NewGroup("/metrics", mids...)
	g2.GET("", router.GoHandlerToEMFHandler(promhttp.Handler()))
}
