package fritzpay

import (
	"database/sql"
	"time"
)

type Payment struct {
	ID        int64
	ProjectID int64
	PaymentID int64
	Created   time.Time
	MethodKey string
}

const (
	TransactionPSPInit  = "psp_init"
	TransactionInit     = "initialized"
	TransactionPSPError = "psp_error"
)

type PaymentTransaction struct {
	FritzpayPaymentID int64
	Timestamp         time.Time
	Status            string
	FritzpayID        sql.NullString
	Payload           sql.NullString
}
