package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
)

// ResponseTemplateWriter implements the http.ResponseWriter interface and allows a go
// template to be executed against the Response body
type ResponseTemplateWriter struct {
	rw       http.ResponseWriter
	template *template.Template
	// response interface{}
}

// Write overrides the Write method by first piping the response data through
// a JSON Decoder and then into the Template
func (tw ResponseTemplateWriter) Write(p []byte) (i int, e error) {
	var (
		pr        *io.PipeReader
		pw        *io.PipeWriter
		errorChan chan error
	)

	errorChan = make(chan error)
	pr, pw = io.Pipe()
	defer func() {
		if err := pw.Close(); err != nil {
			fmt.Printf("I/O Pipe failed to close with error %s", err)
		}
	}()

	// Start reading from the pipe, and return errors via errorChan
	go func() {
		m := map[string]interface{}{}
		if err := json.NewDecoder(pr).Decode(&m); err != nil {
			fmt.Printf("JSON Decoding Error: %s", err)
			errorChan <- err
			return
		}

		fmt.Printf("Executing Template w/ response object \n %+v \n\n", m)

		if err := tw.template.Execute(tw.rw, &m); err != nil {
			fmt.Printf("Template Execution Error: %s", err)
			errorChan <- err
			return
		}

		errorChan <- nil
	}()

	// Write to the input end of the pipe, for reading in the goroutine
	i, e = pw.Write(p)

	// Read errorChan waiting for a result
	if err := <-errorChan; err != nil {
		// Encoding or template execution failed, write the original value instead
		return tw.rw.Write(p)
	}
	return
}

// Header calls the original http ResponseWriter Header method
func (tw ResponseTemplateWriter) Header() http.Header {
	return tw.rw.Header()
}

// WriteHeader calls the original http ResponseWriter WriteHeader method
func (tw ResponseTemplateWriter) WriteHeader(statusCode int) {
	tw.rw.WriteHeader(statusCode)
}

// NewResponseTemplateWriter returns an http.ResponseWriter that executes the given template
// on the http response body, which should be the same type as responseObject
func NewResponseTemplateWriter(
	rw http.ResponseWriter,
	t *template.Template,
	responseObject interface{},
) ResponseTemplateWriter {
	return ResponseTemplateWriter{
		rw:       rw,
		template: t,
		// response: responseObject,
	}
}

// ResponseWrapperMiddleware extends the ResponseTemplateWriter into a Middleware
type ResponseWrapperMiddleware = ResponseTemplateWriter

// Wrapper satisfies the Middleware interface for a ResponseWrapperMiddleware
func (tw ResponseTemplateWriter) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(c echo.Context) (err error) {
		tw.rw = c.Response().Writer
		c.Response().Writer = tw
		return next(c)
	})
}

// WrapResponse is the constructor for a ResponseWrapperMiddleware
func WrapResponse(t *template.Template, responseObject interface{}) ResponseWrapperMiddleware {
	return ResponseWrapperMiddleware{
		template: t,
		// response: responseObject,
	}
}

// TemplateRequest proxies the Request Body from a context through the given template
func TemplateRequest(c echo.Context, t *template.Template) (err error) {
	type objData = map[string]interface{}
	var obj objData

	pr, pw := io.Pipe()

	if err = c.Bind(&obj); err != nil {
		c.Logger().Errorf("Bind error: %s", err)
		if e := pw.CloseWithError(err); e != nil {
			c.Logger().Errorf("Failed to close Pipe with error '%s'", e)
		}
		return
	}
	c.Request().Body = pr

	// Start reading from the pipe, and return errors via errorChan
	go func() {
		fmt.Printf("Executing Template w/ request object \n %+v \n\n", obj)

		if err = t.Execute(pw, obj); err != nil {
			fmt.Printf("Template Execution Error: %s", err)
			if e := pw.CloseWithError(err); e != nil {
				c.Logger().Errorf("Failed to close Pipe with error '%s'", e)
			}
			return
		}
	}()
	return
}

// RequestWrapperMiddleware turns the TemplateRequest function into a Middleware
type RequestWrapperMiddleware struct {
	t *template.Template
}

// Wrapper satisfies the Middleware interface for a RequestWrapperMiddleware
func (rwm RequestWrapperMiddleware) Wrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(c echo.Context) (err error) {
		if err = TemplateRequest(c, rwm.t); err != nil {
			c.Logger().Errorf("New Template Body failed with error: %s", err.Error())
		}
		return next(c)
	})
}

// WrapRequest is the constructor for a RequestWrapperMiddleware
func WrapRequest(t *template.Template) RequestWrapperMiddleware {
	return RequestWrapperMiddleware{t}
}
