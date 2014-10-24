package v1

import (
	"encoding/json"
	"net/http"
)

const (
	StatusImplementationError = "implementationError"
	StatusUnauthorized        = "unauthorized"
	StatusError               = "error"
	StatusSuccess             = "success"
)

const (
	ServiceVersion = "1.1"
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
		ServiceVersion,
		StatusImplementationError,
		"could not read request",
		nil,
		"JSON decoding error",
	}
	ErrUnauthorized = ServiceResponse{
		http.StatusUnauthorized,
		ServiceVersion,
		StatusUnauthorized,
		"unauthorized",
		nil,
		"unauthorized",
	}
	ErrDatabase = ServiceResponse{
		http.StatusInternalServerError,
		ServiceVersion,
		StatusError,
		"database error",
		nil,
		"database error",
	}
	ErrSystem = ServiceResponse{
		http.StatusInternalServerError,
		ServiceVersion,
		StatusError,
		"internal error",
		nil,
		"internal error",
	}
	ErrInval = ServiceResponse{
		http.StatusBadRequest,
		ServiceVersion,
		StatusImplementationError,
		"invalid value",
		nil,
		"invalid value",
	}
	ErrNotFound = ServiceResponse{
		http.StatusNotFound,
		ServiceVersion,
		StatusError,
		"resource not found",
		nil,
		"resource not found",
	}
)

func (sr *ServiceResponse) Write(w http.ResponseWriter) error {

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
