package errors

// TypedError is a simple interface used to expose the ErrorCode from EMFErrors
type TypedError interface {
	ErrorType() string
	Error() string
}

type typedErrorType struct {
	string
}

func (tet *typedErrorType) ErrorType() string {
	return tet.string
}

func (tet *typedErrorType) Error() string {
	return tet.string
}

// ErrorType is the constructor for a TypedError
func ErrorType(code string) (e TypedError) {
	return &typedErrorType{code}
}

// ComponentError is a simple interface used to expose the component that sent a EMFErrror
type ComponentError interface {
	Component() string
	Error() string
}

type componentErrorType struct {
	string
}

func (cet *componentErrorType) Component() string {
	return cet.string
}

func (cet *componentErrorType) Error() string {
	return cet.string
}

// ErrorFromComponent is the constructor for a ComponentError
func ErrorFromComponent(component string) (e ComponentError) {
	return &componentErrorType{component}
}

// ResponseCodeError is a simple interface used to expose the response code from with a EMFErrror ErrorCode
type ResponseCodeError interface {
	Code() int
	Error() string
}

type responseCodeErrorType struct {
	int
}

func (rcet *responseCodeErrorType) Code() int {
	return rcet.int
}

func (rcet *responseCodeErrorType) Error() string {
	return ""
}

// ErrorResponseCode is the constructor for a ResponseCodeError
func ErrorResponseCode(code int) (e ResponseCodeError) {
	return &responseCodeErrorType{code}
}
