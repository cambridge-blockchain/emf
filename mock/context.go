package mock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	"github.com/cambridge-blockchain/emf/emf/context"
)

// ContextType is the EMF Context type that will be sent to each Handler on a request
type ContextType struct {
	context.EMFContext
	t      *testing.T
	req    *http.Request
	rec    *httptest.ResponseRecorder
	v      *viper.Viper
	client *MockClient
}

// Option provides the client a callback that is used to dynamically specify attributes for a
// EMFContext.
type Option func(*ContextType)

// WithClaims is used for specifying the HTTP Client for the Requester to use.
func WithClaims(mapClaims jwt.MapClaims) Option {
	// Use the claims as a map to make linter happy but keep the nice API
	if _, ok := mapClaims["-"]; !ok {
		mapClaims["-"] = ""
	}
	return func(mc *ContextType) {
		mc.Set("user", jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims))
	}
}

// SetConfigMap is used to add fields to the Config object being used for a MockContext
func SetConfigMap(configMap map[string]interface{}) Option {
	return func(mc *ContextType) {
		for key, val := range configMap {
			mc.v.Set(key, val)
		}
	}
}

// NewMockContext is a variadic constructor for a MockContext.
func NewMockContext(
	method string,
	path string,
	body interface{},
	errorsYaml string,
	t *testing.T,
	opts ...Option,
) (mc *ContextType, err error) {
	var (
		r   io.Reader
		w   io.Writer
		ec  echo.Context
		v   *viper.Viper
		req *http.Request
		rec *httptest.ResponseRecorder
	)

	// Pipe the json encoder into the Request input.
	// This allows the caller to send in an arbitrary interface, and everything gets JSON encoded automatically
	r, w = io.Pipe()

	go func() {
		if err = json.NewEncoder(w).Encode(body); err != nil {
			mc.t.Errorf("Could not mock request. JSON Input Encoding Error: %s", err.Error())
		}
	}()

	// Set up the Request Payload
	req = httptest.NewRequest(method, path, r)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderContentEncoding, echo.MIMEApplicationJavaScriptCharsetUTF8)

	// Create a response Recorder
	rec = httptest.NewRecorder()

	// Build a EMF context using an echo context, mock request, and response recorder
	ec = echo.New().NewContext(req, rec)

	v = viper.New()
	v.SetDefault("errors.configPath", errorsYaml)
	v.SetDefault("debug.mode", true)
	// Define component domains here so that the component name is passed through to client.Do
	v.SetDefault("domains.access", "http://access")
	v.SetDefault("domains.admin", "http://admin")
	v.SetDefault("domains.blockchain", "http://blockchain")
	v.SetDefault("domains.keymanagement", "http://keymanagement")
	v.SetDefault("domains.mobile", "http://mobile")
	v.SetDefault("domains.notification", "http://notification")
	v.SetDefault("domains.pds", "http://pds")
	v.SetDefault("domains.serviceprovider", "http://serviceprovider")
	v.SetDefault("domains.schema", "http://schema")
	v.SetDefault("domains.storage", "http://storage")
	v.SetDefault("domains.trustedparty", "http://trustedparty")

	cbc := context.NewEMFContext(ec, v)

	client := &MockClient{t: t, requests: map[string]interface{}{}}
	rh := context.NewRequestHandler(v, ec.Logger(), context.WithHTTPClient(client))
	// t.Logf("Registered Request Handler w/ Storage domain '%s'", rh.GetDomain("storage"))

	context.WithRequestHandler(rh)(cbc)

	// Create the mocked Context object and apply options
	mc = &ContextType{
		cbc,
		t,
		req,
		rec,
		v,
		client,
	}

	for _, opt := range opts {
		opt(mc)
	}
	return
}

// MockRequest is a method to fetch the *http.Request pointer which defines the input request
func (mc *ContextType) MockRequest() *http.Request {
	return mc.req
}

// MockResponse is a method to fetch the ResponseRecorder.Result object to expose the http response
func (mc *ContextType) MockResponse() *http.Response {
	return mc.rec.Result()
}

// MockRecorder is a method to fetch the *httptest.ResponseRecorder pointer which exposes the http response
func (mc *ContextType) MockRecorder() *httptest.ResponseRecorder {
	return mc.rec
}

// DecodeResult is a method to decode the recorded HTTP Response into the provided struct pointer
func (mc *ContextType) DecodeResult(output interface{}) (err error) {
	return json.NewDecoder(mc.MockResponse().Body).Decode(output)
}

// ResultMap returns the recorded HTTP Response Body as a map[string]interface{} and fails the test on Error
func (mc *ContextType) ResultMap() map[string]interface{} {
	type mapPtr map[string]interface{}
	var outMap mapPtr
	if err := json.NewDecoder(mc.MockResponse().Body).Decode(&outMap); err != nil {
		mc.t.Errorf(
			"Could not decode HTTP Response into a map. Result: '%s' \n Error: '%s'",
			mc.MockResponse().Body,
			err,
		)
	}
	return outMap
}

// AddRequestFunction is a method to enable a new Mock Request for this context
func (mc *ContextType) AddRequestFunction(
	method, component, path string,
	// Input to this request function comes directly from the Handler
	// Output can be any value, Mock Requester will re-encode it as needed
	// Send a statusCode over 400 to cause a StatusCode Error
	f func(input interface{}) (statusCode int, output interface{}),
) {
	mc.client.requests[fmt.Sprintf("%s:%s:%s", method, component, path)] = f
}

// AddRequestMap is a method to enable a new Mock Request for this context
func (mc *ContextType) AddRequestMap(
	method, component, path string,
	// Input to this request function comes directly from the Handler
	// Output can be any value, Mock Requester will re-encode it as needed
	// Send a statusCode over 400 to cause a StatusCode Error
	response map[string]interface{},
) {
	mc.client.requests[fmt.Sprintf("%s:%s:%s", method, component, path)] = response
}

// AddRequestInterface is a method to enable a new Mock Request for this context
func (mc *ContextType) AddRequestInterface(
	method, component, path string,
	// Input to this request function comes directly from the Handler
	// Output can be any value, Mock Requester will re-encode it as needed
	// Send a statusCode over 400 to cause a StatusCode Error
	response interface{},
) {
	mc.client.requests[fmt.Sprintf("%s:%s:%s", method, component, path)] = response
}

// AddRequestStatusCode is a method to enable a new Mock Request and facilitate making c.Requester fail
func (mc *ContextType) AddRequestStatusCode(
	method, component, path string, statusCode int,
	// Input to this request function comes directly from the Handler
	// Output can be any value, Mock Requester will re-encode it as needed
	// Send a statusCode over 400 to cause a StatusCode Error
	response interface{},
) {
	mc.client.requests[fmt.Sprintf("%s:%s:%s", method, component, path)] = func(interface{}) (int, interface{}) {
		return statusCode, response
	}
}

// GetRequests is a method to return the map of all Mock Requests for this Context
func (mc *ContextType) GetRequests() map[string]interface{} {
	return mc.client.requests
}

// SetRequests is a method to set the map of all Mock Requests for this Context
func (mc *ContextType) SetRequests(requests map[string]interface{}) {
	mc.client.requests = requests
}

// MockClient is a simple tool to encapsulate the old MockRequest functionality into an http.Client
type MockClient struct {
	requests map[string]interface{}
	t        *testing.T
}

// Do implements the http.Client interface and uses a request map for hard-coded request-to-response mapping
func (client *MockClient) Do(req *http.Request) (res *http.Response, err error) {
	var (
		input   interface{}
		request interface{}
		data    interface{}
		reqID   string
		ok      bool
	)

	res = new(http.Response)
	res.StatusCode = 200

	reqID = fmt.Sprintf("%s:%s:%s", req.Method, req.URL.Hostname(), req.URL.RequestURI())

	// client.t.Logf("Original reqID: '%s', URL: '%s'", reqID, req.URL)

	if request, ok = client.requests[reqID]; !ok {
		reqID = fmt.Sprintf("%s:%s:%s", req.Method, req.Host, req.URL.Path)
		request = client.requests[reqID]
		// client.t.Logf("New reqID: %s, Request: %+v", reqID, request)
	}

	if req.Body != nil {
		if err = json.NewDecoder(req.Body).Decode(&input); err != nil {
			return
		}
	}

	switch val := request.(type) {
	case func(interface{}) (int, interface{}):
		res.StatusCode, data = val(input)
	case func(interface{}) interface{}:
		data = val(input)
	case map[string]interface{}:
		data = val
	case int:
		res.StatusCode = val
	default:
		data = val
	}

	pr, pw := io.Pipe()
	res.Body = pr

	go func() {
		if err = json.NewEncoder(pw).Encode(data); err != nil {
			panic(err)
		}
		if err = pw.Close(); err != nil {
			panic(err)
		}
	}()

	return
}

// Requester : This method is called by the Handler that consumes this MockContext object.
// When the Handler calls Requester, it will pull from the Mock Request map, which is
// manipulated using the AddRequestXXX functions.
func (mc *ContextType) Requester(
	method string,
	component string,
	path string,
	input interface{},
	output interface{},
) (err error) {
	var (
		parsedURL  *url.URL
		request    interface{}
		data       interface{}
		inputBytes []byte
		req        string
		ok         bool
	)
	var statusCode = 200

	if parsedURL, err = url.Parse(path); err != nil {
		// Failing to parse the path suggests that http.NewRequest would have also failed
		return fmt.Errorf("NewRequest Error: %s", err.Error())
	}

	req = fmt.Sprintf("%s:%s:%s", method, component, path)

	if request, ok = mc.client.requests[req]; !ok {
		req = fmt.Sprintf("%s:%s:%s", method, component, parsedURL.Path)
		request = mc.client.requests[req]
	}

	switch val := request.(type) {
	case func(interface{}) (int, interface{}):
		statusCode, data = val(input)
	case func(interface{}) interface{}:
		data = val(input)
	case map[string]interface{}:
		data = val
	case int:
		statusCode = val
	default:
		data = val
	}

	if statusCode >= http.StatusBadRequest {
		return fmt.Errorf("StatusCode Error: %v", statusCode)
	}

	// issues with mapstructure on embedded json struct tags in testing
	// if err = mapstructure.Decode(data, &output); err != nil {
	// 	mc.t.Errorf("Could not decode Mock Requester data into the given pointer. Request: '%s' \n Error: '%s'", req, err)
	// 	return
	// }

	if inputBytes, err = json.Marshal(data); err != nil {
		return
	}

	if err = json.Unmarshal(inputBytes, &output); err != nil {
		return
	}

	// token = mc.Get("user").(*jwt.Token).Raw
	// req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	// req.Header.Add("Content-Type", "application/json")
	// req.Header.Add(echo.HeaderXRequestID, mc.GetRequestID())

	return
}
