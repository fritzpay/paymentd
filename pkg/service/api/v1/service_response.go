package v1

const (
	StatusImplementationError = "implementationError"
	StatusUnauthorized        = "unauthorized"
	StatusError               = "error"
	StatusSuccess             = "success"
)

// ServiceResponse represents a general response container for (payment-related) API
// requests
type ServiceResponse struct {
	Status   string
	Info     string
	Response interface{}
	Error    interface{}
}

// default service responses
var (
	ErrReadJson = ServiceResponse{
		StatusImplementationError,
		"could not read request",
		nil,
		"JSON decoding error",
	}
	ErrUnauthorized = ServiceResponse{
		StatusUnauthorized,
		"unauthorized",
		nil,
		"unauthorized",
	}
	ErrDatabase = ServiceResponse{
		StatusError,
		"database error",
		nil,
		"database error",
	}
	ErrSystem = ServiceResponse{
		StatusError,
		"internal error",
		nil,
		"internal error",
	}
	ErrInval = ServiceResponse{
		StatusImplementationError,
		"invalid value",
		nil,
		"invalid value",
	}
	ErrNotFound = ServiceResponse{
		StatusError,
		"resource not found",
		nil,
		"resource not found",
	}
)
