package paypal_rest

import (
	"time"
)

const (
	TypeSale = "sale"
	TypeAuth = "authorize"
)

type Config struct {
	ProjectID int64
	MethodKey string
	Created   time.Time
	CreatedBy string

	Endpoint string
	ClientID string
	Secret   string
	Type     string
}
