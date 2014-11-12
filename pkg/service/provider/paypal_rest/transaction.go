package paypal_rest

import (
	"database/sql"
	"time"
)

const (
	TransactionTypeCreatePayment         = "createPayment"
	TransactionTypeCreatePaymentResponse = "createPaymentResponse"
)

type Transaction struct {
	ProjectID        int64
	PaymentID        int64
	Timestamp        time.Time
	Type             string
	PaypalID         sql.NullString
	PaypalCreateTime sql.NullString
	PaypalState      sql.NullString
	PaypalUpdateTime *time.Time
	Links            sql.NullString
	Data             sql.NullString
}
