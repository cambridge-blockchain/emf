# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog] and this project adheres to [Semantic Versioning].

## v1.0.0 - 2020-04-15

- Rename to EMF and document accordingly
- First public release to github

## v0.9.5 - 2020-03-09

- Update echo, prometheus, elasticsearch, and datadog dependencies

## v0.9.4 - 2020-02-10

- Typo / error message fixes

## v0.9.3 - 2020-02-07

- Add /noauth/notifications/list to all components
- Add shared creating and sending functionality for notification

## v0.9.2 - 2020-01-16

- Improved distributed tracing, fixes

## v0.9.1 - 2020-01-16

- Improved distributed tracing

## v0.9.0 - 2020-01-15

- Support caching tools, add cache library

## v0.8.1 - 2020-01-09

- Remove param and query param binding via c.Bind to allow for arbitrary json binding

## v0.8.0 - 2020-01-08

- ParamChecker middleware now 404's when params contain '/'
- Upgrade to echo v4.1.13
- Reveal emf and echo versions via /info

## v0.7.6 - 2019-11-04

- Support for errors.Is comparisons
- RequestHandler now supports the Logger() method
- When no exported logging is enabled, don't use logrus

## v0.7.5 - 2019-10-08

- Bug fixes

## v0.7.4 - 2019-09-24

- Disable DataDog logging for metrics and info endpoints
- Support storing request-id for every call in DataDog alongside trace data
- Bug fixes

## v0.7.3 - 2019-08-13

- Revert echo to v0.4.16 due to https://github.com/labstack/echo/issues/1356

## v0.7.2 - 2019-08-13

- Add ParamChecker Middleware
- Make BodyLimitMiddleware more configurable
- Add api.max_http_body config parameter and set default value of 10M

## v0.7.1 - 2019-07-24

- Add DataDog support

## v0.7.0 - 2019-07-08

- Add go modules support and update dependencies accordingly

## v0.6.3 - 2019-06-17

- Add support for c.Validate using the go-playground/validator package

## v0.6.2 - 2019-05-31

- Make the syslog output configurable
- Add support for a configurable max limit

## v0.6.1 - 2019-05-20

- Expose InitRequest through the RequestHandler
- Add GetLimitAndOffset to Context
- Add GetDefaultLimit to RequestHandler
- Add GetRequestLimited helper function to Context (passes through limit and offset parameters)
- Remove ltidFormat references and standardize a new SimpleErrorType
- Add ToSimpleError method to all EMFErrors
- Make requester also handle downstream SimpleErrorTypes
- Have JWT middleware errors return EMFErrors
- Funnel all EMFErrors through the same code to simplify errors when not in Debug Mode

## v0.6.0 - 2019-04-22

- Also disable the logging middleware for /noauth/* paths
- Print RequestID with each request but do not include it in error structs
- Remove RequestID and Claims objects from the errorHandler
- Remove references to .Claims. from the builtin errors
- Remove ltidFormat references and standardize a new SimpleErrorType
- Add ToSimpleError method to all EMFErrors
- Make requester also handle downstream SimpleErrorTypes

## v0.5.1 - 2019-04-19

- Disable the logging middleware for /metrics and /info
- Simplify error message passing code

## v0.5.0 - 2019-02-06

- Move Requester functionality into a variety of tools for handling a request,
	and in a new struct RequestHandler
- Split EMFContextType into incoming-request tools and outgoing-request tools
	(Incoming is just echo.Context + c.GetClaim, everything else is in RequestHandler)
- Improve EMFError handling and error outputs with and without debug mode
- Skip the file size limit for /fattr
- Skip the auth middleware for /metrics
- Update echo to v3.3.10, the final v3 release

## v0.4.11 - 2019-01-23

- Upgrade to echo v3.3.8
- Expose /debug/pprof endpoints correctly

## v0.4.10 - 2018-12-13

- Update error handling code to better interract with downstream errors
- Mock Requester now json decodes output payloads instead of mapstructure
- ErrorHandler never returns StatusCode 0

## v0.4.9 - 2018-10-23

- Update context.Requester(...) to JSON decode an HTTP Request that returns a
	client or server error
- Only return EMFErrors from context.Requester(...)
- Expose models.EMFErrorType to avoid importing the errors package elsewhere

## v0.4.8 - 2018-10-22

- Use a simpler interface for c.NewError
- Add emf.401.TokenInvalidProperty to the builtin errors
- Add support for a template middleware

## v0.4.7 - 2018-10-05

- Fix metrics endpoints

## v0.4.6 - 2018-09-24

- Manually expose Use and Pre methods for UseMiddlewares(router)

## v0.4.5 - 2018-09-24

- Fix interface issues with mock context WithClaims method
- Fix router interface for UseMiddlewares(router)

## v0.4.4 - 2018-09-17

- Fix bug in mock requester bounds check
- Split middleware init and registration into 2 calls
	- Components MUST call baseController.GetMiddlewares().UseMiddlewares(router)
	before starting the server in main.go to register the default middlewares
	- Before calling UseMiddlewares components can replace the default middlewares
	to modify their behaviour / configuration (like enabling an Auth request)
- Allow easy config changes for Mock Requester
- Mocker Requester first looks for a MockRequest with query parameters, and then
	falls back to truncating the path.

## v0.4.3 - 2018-08-15

- Add support for Mock Contexts and a Mock Requester
- Remove old mocking code and rename package to "mock"

## v0.4.2 - 2018-08-10

- Fixed low timeout value (which screwed us up in the last minute of Sprint 14)

## v0.4.1 - 2018-08-08

- Check whether errors.configPath exists on startup
- Fix debug.mode
- Fix c.GetRequestID() function
- Fix bugs in auth middleware error handling

## v0.4.0 - 2018-08-01

- Add error handling package, functions, etc
- Use a better default http client with a shorter timeout
- Add go-playground validate support

## v0.3.1 - 2018-07-26

- Fix segfaults
- Fix float64 support in context.GetClaim

## v0.3.0 - 2018-07-17

- Make CustomContext support domains via environment variables
- Support more types than just string for context.GetClaim
- Add Makefile
- Move logger setup into its own package
- Move logrus support over to an in-house wrapper
- Add support for logging JSON objects with individual fields to elasticsearch
- Update dependencies
- Remove the need for an info-specific middleware, instead use a wrapper function
- Expose middlewares to the calling component for later configuration
- Allow for Auth middleware to support configurable skipping / exclusion
- Clean up prometheus endpoints
- Add tollbooth RateLimitMiddleware (disabled by default)
- Add support for Request ID middleware and retrieving the ID for a given request
- Remove trailing slash from request paths

## v0.2.2 - 2018-06-15

- Refactor public endpoint information

## v0.2.1 - 2018-05-23

- Add Elasticsearch support configured via config.yaml/logging.elasticsearch

## v0.2.0 - 2018-05-04

- Add CustomContext with Requester and GetToken methods
- Add Context Middleware to enable the CustomContext
- Add Token Middleware for providing additional checks via a verifier function, and uses Requester to check the token with the Auth Component
- Use the Context Middleware with every request
- Expose the new Middlewares for per-request use

## v0.1.2 - 2018-05-03

- Remove legacy functionality
- Remove legacy auth and make the claims configurable
- Expose functionality for registering Router Groups with functional middlewares

## v0.1.1 - 2018-04-20

- Fix JWT token signing method
- Expose middleware.User, echo.NewHTTPError, and auth.CustomClaims via models.go
- Remove middleware.User and just use jwt.Token
- Make auth middleware skip /info endpoint and use Auth, Common, and Recover middlewares across ALL endpoints
- Expose RouterGroup functionality

## v0.1.0 - 2018-04-17

- Log to elasticsearch asynchronously
- Move existing emf/endpoint functions to legacy.go
- Move monitor package deeper as part of common middleware and endpoint/monitoring
- Add echo's recover middleware for all endpoints to handle panics
- Use echo's "common" and "auth" middlewares
- Add an info middleware for use in the info endpoint to expose configuration/build info
- Move config into configurer package and remove util
- Refactor config package to include a BuildConfig and Config type
- Make emf.New and server take in the new BuildConfig
- Add ConfigReader type to replace Configurer
- Move mock package as tester with a new constructor
- Introduce models package
- Remove logger from non-legacy applications and use echo/gommon logger instead
- Remove endpoint.go, Registrar, and related types/functions
- Configure echo's logger to be more readable

## v0.0.2 - 2018-04-13

- AddRoute/AddLegacyRoute distinction with echo Contexts
- New Middlewares to support echo Contexts
- Added CHANGELOG.md
- Use AddRoute + echo Contexts for /info
- Log to elasticsearch
- Set the logging level

## v0.0.1 - 2018-04-12

- Supports emf controller class with easy setup
- Monitor, Logger, Router, Server, Middleware, and Endpoint packages
- Use echo for server/routing/middlewares instead of gorilla/mux and custom server/middlewares
- Supports existing endpoints using AddLegacyRoute

# TODO

- [x] The logger might not be configured to use elasticsearch either, but that's straightforward
- [x] There is no router.AddRoute designed to use new-style endpoints with contexts.
- [x] Make middleware interface more generic to allow for both new and legacy routes
- [x] Versions are mostly propogated through but the componentInfo struct requires a ReleaseTimestamp field
- [x] Restructure util to have a better name and maybe more utilities
- [x] The echo server instance itself does not log to elasticsearch or use our logging package at all
- [x] Move mock and other testing tools into emf
- [x] Stress testing to determine how much faster it is
- [x] Remove legacy.go functionality
- [x] Use a custom echo Context to be more flexible in the future https://echo.labstack.com/guide/context
- [x] Add Custom Context functionality for calling out to a component, checking errors, and JSON marshalling into provided type (via xxx-component/models.go)
- [x] Add Custom Context function for extracting the token nicely
- [ ] Find better default behaviour when configFile does not exist
- [ ] CI, Unit Testing, Etc.
- [ ] Convert middlewares to use the new error format
- [ ] Make use of error Levels
- [ ] Make status_code consistent with other fields in the errors.yaml
- [ ] Helper functions for mock requester based on ID

[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/
