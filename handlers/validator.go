package handlers

import "net/http"

// Validator provides methods a to validate a request
type Validator interface {
	Validate(*http.Request, interface{}) error
}
