package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/cambridge-blockchain/emf/emf/context"
	"github.com/cambridge-blockchain/emf/emf/context/errors"
)

func sendJSONorPanic(c echo.Context, code int, isDebug bool, msg interface{}) {
	c.Logger().Error(msg)

	if e, ok := msg.(*errors.EMFErrorType); !isDebug && ok {
		msg = e.ToSimpleError()
	}

	if err := c.JSON(code, msg); err != nil {
		c.Logger().Errorf("Could not return JSON due to error '%s'. Cannot return to client!", err)
	}
}

// CustomHTTPErrorHandler handles errors by both printing them nicely and sending them to elasticsearch
func CustomHTTPErrorHandler(err error, c echo.Context) {
	var code = 500 // Default response code
	var eh = context.NewEMFErrorHandler(
		c,
		c.QueryParam("debug_mode") == "true",
		errors.WithLogger(c.Logger()),
	)

	switch e := err.(type) {
	case *errors.EMFErrorType:
		if e.StatusCode != 0 {
			code = e.StatusCode
		}

		sendJSONorPanic(c, code, eh.DebugMode, e)

	case *echo.HTTPError:

		if e == emiddleware.ErrJWTMissing {
			sendJSONorPanic(c, 400, eh.DebugMode,
				eh.NewError("emf.400.TokenMissing", map[string]interface{}{
					"Error": e,
				}),
			)
			return
		}

		// The only error message that echo actually exposes is ErrJWTMissing
		if e.Message == "invalid or expired jwt" {
			sendJSONorPanic(c, 401, eh.DebugMode,
				eh.NewError("emf.401.TokenExpired", map[string]interface{}{
					"Error": e,
				}),
			)
			return
		}

		if e.Code != 0 {
			code = e.Code
		}
		sendJSONorPanic(c, code, eh.DebugMode, e.Message)

	case jwt.ValidationError:
		sendJSONorPanic(c, 400, eh.DebugMode,
			eh.NewError("emf.400.TokenMissing", map[string]interface{}{
				"Error": e,
			}),
		)

	default:
		sendJSONorPanic(c, code, eh.DebugMode, e)
	}
}
