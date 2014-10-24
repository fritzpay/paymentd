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

// ServiceResponse represents a general response container for (payment-related) API
// requests
type ServiceResponse struct {
	HttpStatus int `json:"-"`
	Status     string
	Info       string
	Response   interface{}
	Error      interface{}
}

// default service responses
var (
	ErrReadJson = ServiceResponse{
		http.StatusBadRequest,
		StatusImplementationError,
		"could not read request",
		nil,
		"JSON decoding error",
	}
	ErrUnauthorized = ServiceResponse{
		http.StatusUnauthorized,
		StatusUnauthorized,
		"unauthorized",
		nil,
		"unauthorized",
	}
	ErrDatabase = ServiceResponse{
		http.StatusInternalServerError,
		StatusError,
		"database error",
		nil,
		"database error",
	}
	ErrSystem = ServiceResponse{
		http.StatusInternalServerError,
		StatusError,
		"internal error",
		nil,
		"internal error",
	}
	ErrInval = ServiceResponse{
		http.StatusBadRequest,
		StatusImplementationError,
		"invalid value",
		nil,
		"invalid value",
	}
	ErrNotFound = ServiceResponse{
		http.StatusNotFound,
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
