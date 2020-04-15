// Package models is intended to expose all types or interfaces that may be needed
// to interract with various emf subpackages, but exclusively in situations where
// functionality from inside these packages is not needed.
// For example, the endpoint package in each EMF service can register routes to a
// router.Group but should not otherwise control routing.
package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/configurer"
	"github.com/cambridge-blockchain/emf/emf/context"
	"github.com/cambridge-blockchain/emf/emf/context/errors"
	"github.com/cambridge-blockchain/emf/emf/middleware"
	"github.com/cambridge-blockchain/emf/emf/router"
)

// MethodPostGob is the HTTP Method string to enable Gob encoding for a POST request
const MethodPostGob = context.MethodPostGob

// MethodPutGob is the HTTP Method string to enable Gob encoding for a POST request
const MethodPutGob = context.MethodPutGob

// UUIDRegex is a regular expression that matches UUIDv4
const UUIDRegex = middleware.UUIDRegex

// User is an internal representation of a user per HTTP request.
type User = *jwt.Token

// HandlerFunc is equivalent to the echo HandlerFunc interface
// to avoid the middleware package having to import it
type HandlerFunc = router.HandlerFunc

// HTTPClient issues HTTP Requests and performs error handling.
type HTTPClient = router.HTTPClient

// Context is the EMF Context method set
type Context = context.EMFContext

// ContextType is the EMF Context Type
type ContextType = context.EMFContextType

// RequestHandler is the subset of a Context used for HTTP Requests
type RequestHandler = context.RequestHandler

// RequestHandlerType is the subset of a Context used for HTTP Requests
type RequestHandlerType = context.RequestHandlerType

// ConfigReader defines the interface for read-only config file access
type ConfigReader = configurer.ConfigReader

// BuildConfig provides the logger and the info endpoint access to hard-coded build configuration
type BuildConfig = configurer.BuildConfig

// Logger is the type of the echo Logger
type Logger = echo.Logger

// Router is the type of the echo Router
type Router = router.Router

// Middleware is the type of our modified echo Middleware
type Middleware = router.Middleware

// RouterGroup is the type of the echo Router Group
type RouterGroup = router.Group

// DataPoint represents a data point value
type DataPoint = configurer.DataPoint

// EMFErrorHandler is the Error Handler Interface exposed by the EMFContext
type EMFErrorHandler = errors.EMFErrorHandler

// EMFError is the Error Interface created by a EMFErrorHandler
type EMFError = errors.EMFError

// TypedError is the TypedError Interface created by a EMFErrorHandler
type TypedError = errors.TypedError

// EMFErrorType is the Error Type created by a EMFErrorHandler
type EMFErrorType = errors.EMFErrorType

// SimpleErrorType is the Simpler Error Type returned by the error_handler middleware
type SimpleErrorType = errors.SimpleErrorType

// ErrorResponseCode is a generic error for use with errors.Is to check against the StatusCode
// 		Example: if errors.Is(err, emf.ErrorResponseCode(401)) {
var ErrorResponseCode = errors.ErrorResponseCode

// ErrorFromComponent is a generic error for use with errors.As to check the error's component name
// 		Example: if errors.Is(err, emf.ErrorFromComponent("kmc")) {
var ErrorFromComponent = errors.ErrorFromComponent

// ErrorType is a generic error for use with error.Is comparisons:
// 		Example: errors.Is(err, emf.ErrorType("sp.400.RequestPayloadInvalid"))
var ErrorType = errors.ErrorType

// WrapRequest is the constructor for a RequestWrapperMiddleware
var WrapRequest = middleware.WrapRequest

// WrapResponse is the constructor for a ResponseWrapperMiddleware
var WrapResponse = middleware.WrapResponse
