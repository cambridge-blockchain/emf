package errors

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/cambridge-blockchain/emf/configurer"
)

// EMFErrorHandler is the Interface which EMFErrorHandlerType implements
type EMFErrorHandler interface {
	NewError(code string, data map[string]interface{}, errors ...error) error
	Logger() echo.Logger
}

// EMFError is the Interface which EMFErrorType implements
type EMFError interface {
	ErrorMap() (errmap map[string]interface{}, err error)
	Execute(eh *EMFErrorHandlerType) (err error)
	ErrorWrap()
	Unwrap() error
	Is() bool
	As() bool
	ToSimpleError() *SimpleErrorType
	GetStackTrace()
	TypedError
}

// EMFErrorHandlerType is the EMF Context type that will be sent to each Handler on a request
type EMFErrorHandlerType struct {
	logger      echo.Logger
	templates   configurer.ConfigReader
	DebugMode   bool
	Method      string
	Path        string
	QueryString string
	Data        map[string]interface{}
}

// EMFErrorType is the type of all EMFErrors handled by the EMFErrorHandler
type EMFErrorType struct {
	Timestamp   string
	StatusCode  int
	ErrorCode   string
	Description string
	Message     map[string]string
	Level       string
	Data        map[string]interface{}
	// Not always Returned
	stackTrace    string // TODO: Should this just wrap the error message itself?
	internalError error
}

// ToSimpleError is a method to convert a EMFErrorType into a SimpleErrorType
func (e EMFErrorType) ToSimpleError() (serr *SimpleErrorType) {
	serr = new(SimpleErrorType)

	serr.Error.Message = e.Error()
	serr.Error.ErrorCode = e.ErrorCode
	serr.StatusCode = e.StatusCode

	return
}

// IsStatusGroup returns a true if the error's status code is in the given block of 100
func (e EMFErrorType) IsStatusGroup(code int) (retVal bool) {
	return ((e.StatusCode / 100) * 100) == code
}

// SimpleErrorType is the type of errors returned to clients outside of debug mode
type SimpleErrorType struct {
	Error struct {
		Message   string `json:"message"`
		ErrorCode string `json:"errorCode"`
	} `json:"error"`
	StatusCode int `json:"statusCode"`
}

// ErrorType is used for loose comparison of EMFErrors with errors.Is and errors.As
func (serr SimpleErrorType) ErrorType() string {
	return serr.Error.ErrorCode
}

// ToEMFError is a method to convert a SimpleErrorType into a EMFErrorType
func (serr SimpleErrorType) ToEMFError() (e *EMFErrorType) {
	e = new(EMFErrorType)

	e.Message = map[string]string{
		"en": serr.Error.Message,
	}
	e.StatusCode = serr.StatusCode
	e.ErrorCode = serr.Error.ErrorCode
	e.Timestamp = time.Now().Format(time.RFC3339)

	return
}

// HandlerOption provides the client a callback that is used to dynamically specify additional EMFErrorHandler options
type HandlerOption func(*EMFErrorHandlerType)

// WithLogger allows the caller to specify the Logger to use for an ErrorHandler
func WithLogger(l echo.Logger) HandlerOption {
	return func(eh *EMFErrorHandlerType) { eh.logger = l }
}

// WithTemplate allows the caller to specify the Error Templates to use for an ErrorHandler
func WithTemplate(path string) HandlerOption {
	return func(eh *EMFErrorHandlerType) {
		eh.templates = configurer.LoadConfig(path, "errors")
	}
}

// Is is a method for comparing errors. It leverages TypedErrors for loose comparisons between EMFErrors
func (e EMFErrorType) Is(target error) bool {
	switch err := target.(type) {
	case TypedError:
		return e.ErrorCode == err.ErrorType()
	case ResponseCodeError:
		return e.StatusCode == err.Code()
	case ComponentError:
		return strings.HasPrefix(e.ErrorCode, err.Component()+".")
	}

	return false
}

// Unwrap is a standard error method for unwrapped the error contained within
func (e EMFErrorType) Unwrap() error {
	return e.internalError
}

// ErrorType implements the TypedError interface and returns the ErrorCode
func (e EMFErrorType) ErrorType() string {
	return e.ErrorCode
}

// Error implements the standard error interface
// TODO: find a better solution for localization
func (e *EMFErrorType) Error() string {
	return e.Message["en"]
}

// Execute executes the Message Templates using the given ErrorHandler struct as input
func (e EMFErrorType) Execute(eh *EMFErrorHandlerType) (err error) {
	for language, message := range e.Message {
		t := template.Must(template.New(language).Parse(message))
		buf := new(bytes.Buffer)
		if err = t.Option("missingkey=zero").ExecuteTemplate(buf, language, eh); err != nil {
			return err
		}
		e.Message[language] = buf.String()
	}
	return
}

// ErrorMap returns a map with all EMFErrorType fields
func (e EMFErrorType) ErrorMap() (errmap map[string]interface{}, err error) {
	if err = mapstructure.Decode(e, &errmap); err != nil {
		return
	}
	return
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ErrorWrap wraps the english error message with a stacktrace, and saves that to stackTrace
func (e *EMFErrorType) ErrorWrap() {
	stack := errors.WithStack(errors.New(e.Message["en"])).(stackTracer)
	frames := stack.StackTrace()
	minimumStackLength := 9
	if len(frames) > minimumStackLength {
		e.stackTrace = fmt.Sprintf("%+v", frames[3:len(frames)-6])
	} else {
		e.stackTrace = fmt.Sprintf("%+v", frames)
	}
}

// GetStackTrace returns the hidden stackTrace field, or sets it if uninitialized
func (e EMFErrorType) GetStackTrace() string {
	if e.stackTrace == "" {
		e.ErrorWrap()
	}
	return e.stackTrace
}

// Logger returns the embedded echo.Logger
func (eh EMFErrorHandlerType) Logger() echo.Logger {
	return eh.logger
}

// NewError is a method used to generate and log an EMFError message using the configured template
func (eh *EMFErrorHandlerType) NewError(code string, data map[string]interface{}, errors ...error) error {
	var (
		ok       bool
		err      error
		errorMap map[string]interface{}
		e        EMFErrorType
	)

	if e, ok = getBuiltin(code); !ok {
		if err = eh.templates.UnmarshalKey(code, &e); err != nil {
			return err
		}
	}

	e.Timestamp = time.Now().Format(time.RFC3339)
	e.ErrorCode = code

	// Ignores any Internal Errors after the first
	if len(errors) > 0 {
		e.internalError = errors[0]
		if dataErr, ok := data["Error"]; !ok || dataErr == nil {
			data["Error"] = errors[0]
		}
	} else if dataErr, ok := data["Error"].(error); ok {
		e.internalError = dataErr
	}

	if eh.DebugMode {
		if err = mapstructure.Decode(eh, &data); err != nil {
			return err
		}
	}
	e.Data = data
	eh.Data = data

	if err = e.Execute(eh); err != nil {
		return err
	}

	if errorMap, err = e.ErrorMap(); err != nil {
		return err
	}

	e.ErrorWrap()
	if eh.DebugMode {
		eh.logger.Debug(e.GetStackTrace())
	}
	eh.logger.Debugj(errorMap)

	return &e
}
