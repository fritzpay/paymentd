package v1

import (
	"encoding/json"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

const (
	StatusImplementationError = "implementationError"
	StatusUnauthorized        = "unauthorized"
	StatusError               = "error"
	StatusSuccess             = "success"
)

const (
	// APIVersion is the current version of the API
	//
	// Version history:
	//
	//   - 1.2: Deprecating "Error" field. Will be removed in version 2
	//
	//   - 1.1: Include version number in service response
	APIVersion = "1.2"
)

// ServiceResponse represents a general response container for (payment-related) API
// requests
type ServiceResponse struct {
	HttpStatus int `json:"-"`
	Version    string
	Status     string
	Info       string
	Response   interface{}
	Error      interface{}
}

// default service responses
var (
	ErrReadJson = ServiceResponse{
		http.StatusBadRequest,
		APIVersion,
		StatusImplementationError,
		"could not read request",
		nil,
		nil,
	}
	ErrUnauthorized = ServiceResponse{
		http.StatusUnauthorized,
		APIVersion,
		StatusUnauthorized,
		"unauthorized",
		nil,
		nil,
	}
	ErrDatabase = ServiceResponse{
		http.StatusInternalServerError,
		APIVersion,
		StatusError,
		"database error",
		nil,
		nil,
	}
	ErrSystem = ServiceResponse{
		http.StatusInternalServerError,
		APIVersion,
		StatusError,
		"internal error",
		nil,
		nil,
	}
	ErrInval = ServiceResponse{
		http.StatusBadRequest,
		APIVersion,
		StatusImplementationError,
		"invalid value",
		nil,
		nil,
	}
	ErrNotFound = ServiceResponse{
		http.StatusNotFound,
		APIVersion,
		StatusError,
		"resource not found",
		nil,
		nil,
	}
	ErrConflict = ServiceResponse{
		http.StatusConflict,
		APIVersion,
		StatusError,
		"resource already exits",
		nil,
		nil,
	}
	ErrReadParam = ServiceResponse{
		http.StatusBadRequest,
		APIVersion,
		StatusError,
		"parameter malformed",
		nil,
		nil,
	}
	ErrMethod = ServiceResponse{
		http.StatusMethodNotAllowed,
		APIVersion,
		StatusError,
		"method not allowed",
		nil,
		nil,
	}
)

func (sr *ServiceResponse) Write(w http.ResponseWriter) error {
	if sr.Error != nil {
		Log.Warn("use of deprecated Error field", log15.Ctx{"ServiceResponse": sr})
	}

	// set default http states
	if sr.HttpStatus == 0 && sr.Status == StatusSuccess && sr.Error == nil {
		sr.HttpStatus = http.StatusOK
	} else if sr.HttpStatus == 0 {
		sr.HttpStatus = http.StatusInternalServerError
	}

	w.WriteHeader(sr.HttpStatus)

	// json encode response struct
	je := json.NewEncoder(w)
	err := je.Encode(sr)
	if err != nil {
		return err
	}

	return err
}
