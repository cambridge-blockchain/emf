package logger

import (
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

// This file exists to wrap the echo logger and provide a solid middleware.
// Based directly on https://github.com/sandalwing/echo-logrusmiddleware (unmaintained)

const base10 = 10

// Logger is the logrus Logger assigned globally for this package
type Logger struct {
	*logrus.Logger
}

// Level is an echo logger wrapper for getting the log Level
func (l Logger) Level() log.Lvl {
	switch l.Logger.Level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	case logrus.InfoLevel:
		return log.INFO
	default:
		l.Panic("Invalid level")
	}

	return log.OFF
}

// SetLevel is an echo logger wrapper for setting the Log Level
func (l Logger) SetLevel(lvl log.Lvl) {
	switch lvl {
	case log.DEBUG:
		logrus.SetLevel(logrus.DebugLevel)
	case log.WARN:
		logrus.SetLevel(logrus.WarnLevel)
	case log.ERROR:
		logrus.SetLevel(logrus.ErrorLevel)
	case log.INFO:
		logrus.SetLevel(logrus.InfoLevel)
	default:
		l.Panic("Invalid level")
	}
}

// Output is an echo logger wrapper for getting the raw Output
func (l Logger) Output() io.Writer {
	return l.Out
}

// SetOutput is an echo logger wrapper for setting the Output
func (l Logger) SetOutput(w io.Writer) {
	logrus.SetOutput(w)
}

// Printj is a function for logging a JSON object to logrus, with Fields
func (l Logger) Printj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Print()
}

// Debugj is a function for logging a JSON object to logrus, with Fields
func (l Logger) Debugj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Debug()
}

// Infoj is a function for logging a JSON object to logrus, with Fields
func (l Logger) Infoj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Info()
}

// Warnj is a function for logging a JSON object to logrus, with Fields
func (l Logger) Warnj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Warn()
}

// Errorj is a function for logging a JSON object to logrus, with Fields
func (l Logger) Errorj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Error()
}

// Fatalj is a function for logging a JSON object to logrus, with Fields
func (l Logger) Fatalj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Fatal()
}

// Panicj is a function for logging a JSON object to logrus, with Fields, then Panicking
func (l Logger) Panicj(j log.JSON) {
	logrus.WithFields(logrus.Fields(j)).Panic()
}

// SetHeader is an echo logger wrapper for setting a prefix. Not supported by logrus.
func (l Logger) SetHeader(s string) {
	// TODO.  Could maybe be used to manipulate the logrus Formatter interface
}

// SetPrefix is an echo logger wrapper for setting a prefix. Not supported by logrus.
func (l Logger) SetPrefix(s string) {
	// TODO.  Could maybe be used to manipulate the logrus Formatter interface
}

// Prefix is an echo logger wrapper for getting a prefix. Not supported by logrus.
func (l Logger) Prefix() string {
	// TODO.  Could maybe be used to manipulate the logrus Formatter interface
	return ""
}

func logrusMiddlewareHandler(c echo.Context, next echo.HandlerFunc) error {
	req := c.Request()
	res := c.Response()
	start := time.Now()
	if err := next(c); err != nil {
		c.Error(err)
	}
	stop := time.Now()

	p := req.URL.Path
	if p == "" {
		p = "/"
	}

	bytesIn := req.Header.Get(echo.HeaderContentLength)
	if bytesIn == "" {
		bytesIn = "0"
	}

	logrus.WithFields(map[string]interface{}{
		"request_id":    req.Header.Get(echo.HeaderXRequestID),
		"time_rfc3339":  time.Now().Format(time.RFC3339),
		"remote_ip":     c.RealIP(),
		"host":          req.Host,
		"uri":           req.RequestURI,
		"method":        req.Method,
		"path":          p,
		"referer":       req.Referer(),
		"user_agent":    req.UserAgent(),
		"status":        res.Status,
		"latency":       strconv.FormatInt(int64(stop.Sub(start).Seconds()), base10),
		"latency_human": stop.Sub(start).String(),
		"bytes_in":      bytesIn,
		"bytes_out":     strconv.FormatInt(res.Size, base10),
	}).Info("Handled request")

	return nil
}

func logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return logrusMiddlewareHandler(c, next)
	}
}

// Middleware exposes a logging MiddlewareFunc for echo
func Middleware() echo.MiddlewareFunc {
	return logger
}
