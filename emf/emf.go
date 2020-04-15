package emf

import (
	"fmt"
	"os"
	"time"

	"github.com/cambridge-blockchain/emf/notifications"

	"github.com/labstack/echo/v4"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/cambridge-blockchain/emf/configurer"
	"github.com/cambridge-blockchain/emf/emf/bind"
	"github.com/cambridge-blockchain/emf/emf/endpoint"
	"github.com/cambridge-blockchain/emf/emf/logger"
	"github.com/cambridge-blockchain/emf/emf/middleware"
	"github.com/cambridge-blockchain/emf/emf/router"
	"github.com/cambridge-blockchain/emf/emf/server"
)

// Constants for configuring the server
const (
	headerBytes int = 1 << 16
	// Set according to https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	readTimeout  time.Duration = 120 * time.Second
	writeTimeout time.Duration = 120 * time.Second
	idleTimeout  time.Duration = 120 * time.Second

	Version string = "v1.0.0"
)

// Controller : Contains all of the necessary components to configure and run
// a Cambridge Blockchain microservice api
type Controller struct {
	server      *server.Server
	router      *router.Router
	config      configurer.ConfigReader
	build       configurer.BuildConfig
	middlewares *middleware.AllMiddlewares
}

// GetBuild is a method to expose the config
func (c *Controller) GetBuild() configurer.BuildConfig {
	return c.build
}

// GetConfig is a method to expose the config
func (c *Controller) GetConfig() configurer.ConfigReader {
	return c.config
}

// GetServer is a method to expose the logger
func (c *Controller) GetServer() *server.Server {
	return c.server
}

// GetLogger is a method to expose the logger
func (c *Controller) GetLogger() echo.Logger {
	return c.router.Logger
}

// GetRouter is a method to expose the router
func (c *Controller) GetRouter() *router.Router {
	return c.router
}

// GetMiddlewares is a method to expose the middlewares struct
func (c *Controller) GetMiddlewares() *middleware.AllMiddlewares {
	return c.middlewares
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// New : Creates a default EMF server object, properly configured
func New(configFile string, buildConfig configurer.BuildConfig, notificationCodes []notifications.NotificationType) (
	c *Controller) {
	var (
		s    *server.Server
		r    *router.Router
		e    *echo.Echo
		m    *middleware.AllMiddlewares
		conf configurer.Config
	)

	// ***********************************************
	// * Load up the config
	// ***********************************************

	conf = configurer.LoadConfig(configFile, "emf")
	buildConfig.Component = conf.GetString("name")
	buildConfig.EMFVersion = Version
	buildConfig.EchoVersion = echo.Version

	if _, err := os.Stat(os.ExpandEnv(conf.GetString("errors.configPath"))); os.IsNotExist(err) {
		panic(fmt.Errorf(
			"failed to start server, the configured errors.configPath file '%s' does not exist",
			err.(*os.PathError).Path,
		))
	}

	// ***********************************************
	// * Start up Echo
	// ***********************************************

	e = echo.New()

	// ***********************************************
	// * Set up Logger
	// ***********************************************

	e.Logger = logger.New(
		conf,
		buildConfig,
	)

	// ***********************************************
	// * Configure Debug Mode
	// ***********************************************
	if conf.GetBool("debug.mode") {
		e.Logger.Info("Using debug mode for more verbose output...")
		e.Debug = true
	}

	// ***********************************************
	// * Set up Server
	// ***********************************************

	e.Server.Addr = fmt.Sprintf(":%v", conf.GetString("api.port"))
	e.Server.ReadTimeout = readTimeout
	e.Server.WriteTimeout = writeTimeout
	e.Server.IdleTimeout = idleTimeout
	e.Server.MaxHeaderBytes = headerBytes

	// Use our configered echo Server
	s = server.New(server.WithServer(e.Server))

	// Register Custom HTTP Error Handler for EMFErrors
	e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler

	// Register Custom Validator based on go-playground validator
	e.Validator = &customValidator{validator: validator.New()}

	// ***********************************************
	// * FIX BIND
	// ***********************************************

	e.Binder = &bind.DefaultBinder{}

	// ***********************************************
	// * Expose Middlewares
	// ***********************************************

	m = middleware.InitMiddlewares(conf)

	// ***********************************************
	// * Set up router and register Routes
	// ***********************************************

	r = router.New(router.WithRouter(e))

	endpoint.RegisterInfo(r, buildConfig)
	endpoint.RegisterNotification(r, notificationCodes)

	// ***********************************************
	// * Configure performance monitoring
	// ***********************************************

	if conf.GetBool("monitoring.prometheus") {
		endpoint.RegisterMonitoring(r)
	}

	// ***********************************************
	// * Return the controller
	// ***********************************************

	c = &Controller{
		server:      s,
		router:      r,
		config:      conf,
		build:       buildConfig,
		middlewares: m,
	}

	return
}
