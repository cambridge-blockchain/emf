package context

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/cambridge-blockchain/emf/configurer"
	"github.com/cambridge-blockchain/emf/emf/context/errors"
)

// RequestHandlerType is the minimum struct for sending requests with Requester
type RequestHandlerType struct {
	cfg    configurer.ConfigReader
	client Client
	eh     *errors.EMFErrorHandlerType
	header http.Header
}

// RequestHandler is the minimum method set for the Requester family of functions
type RequestHandler interface {
	Logger() echo.Logger
	LoggerlessRequestHandler
}

// LoggerlessRequestHandler is the minimum method set for the Requester family of functions
type LoggerlessRequestHandler interface {
	Requester(method, component, path string, input, output interface{}) error
	InitRequest(
		encodingType string, method string, domain string, path string, input interface{},
	) (req *http.Request, err error)
	GetRequest(component, path string) (*http.Request, error)
	JSONRequest(method, component, path string, input interface{}) (*http.Request, error)
	GobRequest(method, component, path string, input interface{}) (*http.Request, error)
	SendRequest(request *http.Request, output interface{}) error
	Header() http.Header
	GetDomain(component string) (domain string)
	IsDebug() bool
	ErrorHandler() errors.EMFErrorHandler
	NewError(code string, data map[string]interface{}, errors ...error) error
	GetDefaultLimit() (limit int)
	GetMaxLimit() (limit int)
}

// RHOption provides the client a callback that is used to dynamically specify attributes for a
// RequestHandler.
type RHOption func(*RequestHandlerType)

// Client is an interface that mirrors the http.Client interface
type Client interface {
	Do(req *http.Request) (res *http.Response, err error)
}

// WithHTTPClient is used for specifying the HTTP Client for the Requester to use.
func WithHTTPClient(client Client) RHOption {
	return func(rh *RequestHandlerType) { rh.client = client }
}

// WithDebugMode enables debug mode on the RequestHandler
func WithDebugMode() RHOption {
	return func(rh *RequestHandlerType) { rh.eh.DebugMode = true }
}

// WithHeaders configures the Authorization and other HTTP headers for a requestHandler
func WithHeaders(headers http.Header) RHOption {
	return func(rh *RequestHandlerType) {
		rh.header = headers
	}
}

// NewRequestHandler is a variadic constructor for a RequestHandler.
func NewRequestHandler(cfg configurer.ConfigReader, logger echo.Logger, opts ...RHOption) (rh *RequestHandlerType) {
	rh = &RequestHandlerType{
		cfg,
		&http.Client{},
		&errors.EMFErrorHandlerType{
			DebugMode: false,
		},
		http.Header{},
	}

	errors.WithLogger(logger)(rh.eh)
	errors.WithTemplate(cfg.GetString("errors.configPath"))(rh.eh)

	for _, opt := range opts {
		opt(rh)
	}
	return
}

// ErrorHandler is a helper function to make an ErrorHandler object from the Context and return it
func (rh RequestHandlerType) ErrorHandler() errors.EMFErrorHandler {
	return rh.eh
}

// NewError is a helper function to wrap ErrorHandler.NewError
func (rh RequestHandlerType) NewError(code string, data map[string]interface{}, errors ...error) error {
	return rh.ErrorHandler().NewError(code, data, errors...)
}

// Logger is a helper function to wrap ErrorHandler.Logger
func (rh RequestHandlerType) Logger() echo.Logger {
	return rh.ErrorHandler().Logger()
}

// IsDebug is a helper function to provide access to the Debug Mode config flag
func (rh RequestHandlerType) IsDebug() bool {
	return rh.cfg.GetBool("debug.mode") || (rh.eh != nil && rh.eh.DebugMode)
}

// GetDomain is a helper function to expose fetching domains from the config file
func (rh RequestHandlerType) GetDomain(component string) (domain string) {
	return rh.cfg.GetString("domains." + component)
}

// GetMaxLimit is a helper function to expose fetching the maximum limit from the config file
func (rh RequestHandlerType) GetMaxLimit() (limit int) {
	limit = rh.cfg.GetInt("api.max_limit")
	if limit == 0 {
		limit = 50
	}
	return limit
}

// GetDefaultLimit is a helper function to expose fetching the default limit from the config file
func (rh RequestHandlerType) GetDefaultLimit() (limit int) {
	limit = rh.cfg.GetInt("api.default_limit")
	if limit == 0 {
		limit = 20
	}
	return limit
}

// Header is a helper function to expose the http.Header object
func (rh RequestHandlerType) Header() http.Header {
	return rh.header
}

// DecodeEMFError is a helper function to read a request body and parse it into a EMFErrorType object
// or return an error, also as a EMFErrorType
func DecodeEMFError(r io.Reader, eh errors.EMFErrorHandler) (e *errors.EMFErrorType) {
	var err error

	var body []byte

	// Save the body for re-use later
	if body, err = ioutil.ReadAll(r); err != nil {
		return eh.NewError("emf.500.RequesterErrorResponseFailure", map[string]interface{}{
			"Response": body,
			"Error":    err,
		}).(*errors.EMFErrorType)
	}

	// Try EMFErrorType
	if err = json.NewDecoder(
		ioutil.NopCloser(bytes.NewBuffer(body)),
	).Decode(&e); err == nil && e.ErrorCode != "" {
		return
	}

	// Try a SimpleErrorType
	var serr errors.SimpleErrorType
	if err = json.NewDecoder(
		ioutil.NopCloser(bytes.NewBuffer(body)),
	).Decode(&serr); err == nil && serr.Error.ErrorCode != "" {
		return serr.ToEMFError()
	}

	// Fail. Use interface{} and wrap it in a EMFErrorType
	var badError interface{}

	err = json.NewDecoder(
		ioutil.NopCloser(bytes.NewBuffer(body)),
	).Decode(&badError)

	return eh.NewError("emf.500.RequesterErrorResponseFailure", map[string]interface{}{
		"Response": badError,
		"Error":    err,
	}).(*errors.EMFErrorType)
}

// EncodingJSON is the JSON Input encoding string for c.Requester
const EncodingJSON = "JSON"

// EncodingGob is the Gob Input encoding string for c.Requester
const EncodingGob = "Gob"

// MethodPostGob is the HTTP Method string to enable Gob encoding for a POST request
const MethodPostGob = "POST-GOB"

// MethodPutGob is the HTTP Method string to enable Gob encoding for a POST request
const MethodPutGob = "PUT-GOB"

func encodeInput(encodingType string, input interface{}) (body io.Reader, contentType string, err error) {
	var (
		serializer bytes.Buffer
	)

	body = &serializer

	switch encodingType {
	case EncodingJSON:
		contentType = "application/json"
		if input == nil {
			return
		}
		if err = json.NewEncoder(&serializer).Encode(input); err != nil {
			return
		}

		return
	case EncodingGob:
		contentType = "application/x-gob"
		if input == nil {
			return
		}
		if err = gob.NewEncoder(&serializer).Encode(&input); err != nil {
			return
		}
		return
	default:
		err = fmt.Errorf("invalid encoding type specified '%s'", encodingType)
		return
	}
}

func streamCloser(body io.Closer) {
	var err error
	if body != nil {
		if err = body.Close(); err != nil {
			panic(err)
		}
	}
}

// Requester is a method to send HTTP client requests to other components
// Uses the same interface as before, now with modular request generation
func (rh RequestHandlerType) Requester(
	method string,
	component string,
	path string,
	input interface{},
	output interface{},
) (err error) {
	// Get and Check the Domain
	var domain string
	if domain = rh.GetDomain(component); domain == "" {
		return rh.NewError("emf.500.RequesterCreateRequestFailure", map[string]interface{}{
			"Error": fmt.Errorf("get Domain Error: component '%s' was not configured", component),
		})
	}

	var req *http.Request

	switch method {
	case http.MethodGet:
		if req, err = rh.GetRequest(domain, path); err != nil {
			return
		}
	case MethodPostGob:
		if req, err = rh.GobRequest(http.MethodPost, domain, path, input); err != nil {
			return
		}
	case MethodPutGob:
		if req, err = rh.GobRequest(http.MethodPut, domain, path, input); err != nil {
			return
		}
	default:
		if req, err = rh.JSONRequest(method, domain, path, input); err != nil {
			return
		}
	}

	return rh.SendRequest(req, output)
}

// JSONRequest is the method to set up a JSON-encoded HTTP Request
func (rh RequestHandlerType) JSONRequest(
	method, domain, path string,
	input interface{},
) (req *http.Request, err error) {
	return rh.InitRequest(EncodingJSON, method, domain, path, input)
}

// GobRequest is the method to set up a Gob-encoded HTTP Request
func (rh RequestHandlerType) GobRequest(
	method, domain, path string,
	input interface{},
) (req *http.Request, err error) {
	return rh.InitRequest(EncodingGob, method, domain, path, input)
}

// GetRequest is the method to set up a HTTP Get request with no body
func (rh RequestHandlerType) GetRequest(domain, path string) (req *http.Request, err error) {
	return rh.InitRequest(EncodingGob, http.MethodGet, domain, path, nil)
}

// InitRequest is the method to initialize any HTTP request for use with rh.SendRequest
func (rh RequestHandlerType) InitRequest(
	encodingType string,
	method string,
	domain string,
	path string,
	input interface{},
) (req *http.Request, err error) {
	var (
		requestBody io.Reader
		contentType string
	)

	// Encode Input (if nil, do nothing)
	if input != nil {
		if requestBody, contentType, err = encodeInput(encodingType, input); err != nil {
			return nil, rh.NewError("emf.500.RequesterEncodingFailure", map[string]interface{}{
				"Error": err,
			})
		}
	} else {
		contentType = "application/json"
	}

	// If we're in Debug Mode, pass the debug_mode parameter through to the next request
	if rh.IsDebug() {
		var p *url.URL
		if p, err = url.Parse(path); err == nil {
			p.Query().Add("debug_mode", "true")
			path = p.String()
		}
	}

	if req, err = http.NewRequest(method, domain+path, requestBody); err != nil {
		return nil, rh.NewError("emf.500.RequesterCreateRequestFailure", map[string]interface{}{
			"Error": err,
		})
	}

	// Add headers from the RequestHandler's list
	for k, vals := range rh.header {
		for _, val := range vals {
			req.Header.Add(k, val)
		}
	}

	// Add default request headers
	req.Header.Add(echo.HeaderContentType, contentType)
	req.Header.Add(echo.HeaderContentEncoding, echo.MIMEApplicationJavaScriptCharsetUTF8)

	return
}

// SendRequest is a method to send HTTP client requests to other components
func (rh RequestHandlerType) SendRequest(
	req *http.Request,
	output interface{},
) (err error) {
	// Send the Request
	var res *http.Response
	if res, err = rh.client.Do(req); err != nil {
		return rh.NewError("emf.500.RequesterSendRequestFailure", map[string]interface{}{
			"Error": err,
		})
	}
	defer streamCloser(res.Body)

	// Handle Errors
	if res.StatusCode >= http.StatusBadRequest {
		return DecodeEMFError(res.Body, rh.ErrorHandler())
	}

	// Decode Response
	if err = json.NewDecoder(res.Body).Decode(output); err != nil {
		return rh.NewError("emf.500.RequesterDecodingFailure", map[string]interface{}{
			"Error": err,
		})
	}

	return
}
