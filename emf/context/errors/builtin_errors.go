package errors

import "net/http"

// nolint: lll
func getBuiltin(code string) (cbe EMFErrorType, ok bool) {
	cbe, ok = map[string]EMFErrorType{
		"emf.400.QueryParameterInvalid": {
			ErrorCode:   "emf.400.QueryParameterInvalid",
			StatusCode:  http.StatusBadRequest,
			Description: "An invalid query parameter was provided in the request.",
			Message: map[string]string{
				"en": "The incoming request payload has an invalid query parameter '{{.Data.Param}}'. Error: '{{.Data.Error}}'",
			},
			Data: map[string]interface{}{
				"Param": "The invalid Query parameter.",
				"Error": "The error raised while validating the query parameters.",
			},
		},
		"emf.400.TokenMissing": {
			ErrorCode:   "emf.400.TokenMissing",
			StatusCode:  http.StatusBadRequest,
			Description: "The JWT Token Verification request failed, so the API Request could not be completed.",
			Message: map[string]string{
				"en": `JWT Token is missing or malformed and could not be validated. Error: '{{.Data.Error}}'`,
			},
			Data: map[string]interface{}{
				"Error": "Error found when extracting the JWT token from the request.",
			},
		},
		"emf.400.InvalidParametersFailure": {
			ErrorCode:   "emf.400.InvalidParametersFailure",
			StatusCode:  http.StatusBadRequest,
			Description: "The parameters are invalid.",
			Data: map[string]interface{}{
				"Error": "The error raised while parsing the request parameters.",
			},
			Message: map[string]string{
				"en": "The request parameters are invalid. Error: '{{.Data.Error}}'",
			},
		},
		"emf.401.Unauthorized": {
			ErrorCode:   "emf.401.Unauthorized",
			StatusCode:  http.StatusUnauthorized,
			Description: "The API request cannot be performed as the User associated with the JWT token in the request does not have access to this endpoint.",
			Message: map[string]string{
				"en": "User with identifier '{{.Data.Target}}' and role '{{.Data.Role}}' does not have access.",
			},
			Data: map[string]interface{}{
				"Role":   "Type of authorization access being used.",
				"Target": "Identity being acted on.",
			},
		},
		"emf.401.TokenExpired": {
			ErrorCode:   "emf.401.TokenExpired",
			StatusCode:  http.StatusUnauthorized,
			Description: "The JWT Token provided with the request is Expired.",
			Message: map[string]string{
				"en": "The provided access token expired at '{{.Data.ExpirationDate}}'.",
			},
			Data: map[string]interface{}{
				"ExpirationDate": "Expiration Date of the given token.",
			},
		},
		"emf.401.TokenInactive": {
			ErrorCode:   "emf.401.TokenInactive",
			StatusCode:  http.StatusUnauthorized,
			Description: "The JWT Token provided with the request is no longer Active (the User logged out or switched roles).",
			Message: map[string]string{
				"en": "The provided token with User '{{.Data.Target}}' and role '{{.Data.Role}}' is no longer active.",
			},
			Data: map[string]interface{}{
				"Role":   "Type of authorization access being used.",
				"Target": "Identity being acted on.",
			},
		},
		"emf.401.TokenVerificationFailure": {
			ErrorCode:   "emf.401.TokenVerificationFailure",
			StatusCode:  http.StatusUnauthorized,
			Description: "The JWT Token Verification request failed, so the API Request could not be completed.",
			Message: map[string]string{
				"en": `The provided token with User '{{.Data.Target}}' and role '{{.Data.Role}}' could not be verified,
please try again. Error: '{{.Data.error}}'`,
			},
			Data: map[string]interface{}{
				"Role":   "Type of authorization access being used.",
				"Target": "Identity being acted on.",
			},
		},
		"emf.401.TokenInvalidProperty": {
			ErrorCode:   "emf.401.TokenInvalidProperty",
			StatusCode:  http.StatusUnauthorized,
			Description: "The JWT Token in the request has an invalid property, so the API Request could not be completed. Please try again.",
			Message: map[string]string{
				"en": `The provided token has an invalid '{{.Data.Name}}' value
				'{{.Data.Value}}', please try again. Error: '{{.Data.Error}}'`,
			},
			Data: map[string]interface{}{
				"Error": "Error found when retrieving the property from the JWT token.",
				"Name":  "Name of the JWT token property.",
				"Value": "Value associated with JWT token property.",
			},
		},
		"emf.401.UnauthorizedCaller": {
			ErrorCode:   "emf.401.UnauthorizedCaller",
			StatusCode:  http.StatusUnauthorized,
			Description: "The API request sent by the Caller/Client cannot be performed as the User associated with the JWT token in the API request does not have access to this endpoint.",
			Message: map[string]string{
				"en": "User with identifier '{{.Data.Target}}' and role '{{.Data.Role}}' does not have access.",
			},
			Data: map[string]interface{}{
				"Role":   "Type of authorization access being used.",
				"Target": "Identity being acted on.",
			},
		},
		"emf.500.RequesterEncodingFailure": {
			ErrorCode:   "emf.500.RequesterEncodingFailure",
			StatusCode:  http.StatusInternalServerError,
			Description: "The internal HTTP Request could not be performed as the request or response could not be JSON encoded",
			Message: map[string]string{
				"en": "Failed to encode request payload. Error: '{{.Data.Error}}'",
			},
			Data: map[string]interface{}{
				"Error": "JSON Encoding Error",
			},
		},
		"emf.500.RequesterDecodingFailure": {
			ErrorCode:   "emf.500.RequesterDecodingFailure",
			StatusCode:  http.StatusInternalServerError,
			Description: "The internal HTTP Request could not be performed as the response could not be JSON decoded",
			Message: map[string]string{
				"en": "Failed to decode response payload. Error: '{{.Data.Error}}'",
			},
			Data: map[string]interface{}{
				"Error": "JSON Decoding Error",
			},
		},
		"emf.500.RequesterErrorResponseFailure": {
			ErrorCode:   "emf.500.RequesterErrorResponseFailure",
			StatusCode:  http.StatusInternalServerError,
			Description: "The internal HTTP Request could not be performed as the CBError response could not be JSON decoded",
			Message: map[string]string{
				"en": "Failed to decode failing response payload as CBError. Response: '{{.Data.Response}}'",
			},
			Data: map[string]interface{}{
				"Response": "Response sent by the downstream component",
				"Error":    "Error encountered while trying to decode 'Response'",
			},
		},
		"emf.500.RequesterCreateRequestFailure": {
			ErrorCode:   "emf.500.RequesterCreateRequestFailure",
			StatusCode:  http.StatusInternalServerError,
			Description: "The internal HTTP Request could not be performed as the request object could not be initialized",
			Message: map[string]string{
				"en": "Failed to initialize request. Error: '{{.Data.Error}}'",
			},
			Data: map[string]interface{}{
				"Error": "Error created while initializing the http request",
			},
		},
		"emf.500.RequesterSendRequestFailure": {
			ErrorCode:   "emf.500.RequesterSendRequestFailure",
			StatusCode:  http.StatusInternalServerError,
			Description: "The internal HTTP Request could not be performed as the HTTP client failed to execute the request",
			Message: map[string]string{
				"en": "Failed to execute request. Error: '{{.Data.Error}}'",
			},
			Data: map[string]interface{}{
				"Error": "HTTP client error",
			},
		},
	}[code]
	return
}
