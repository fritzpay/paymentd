package paypal_rest

import (
	"database/sql"
	"time"
)

const (
	TransactionTypeCreatePayment         = "createPayment"
	TransactionTypeCreatePaymentResponse = "createPaymentResponse"
	TransactionTypeError                 = "error"
)

type Transaction struct {
	ProjectID        int64
	PaymentID        int64
	Timestamp        time.Time
	Type             string
	PaypalID         sql.NullString
	PaypalCreateTime *time.Time
	PaypalState      sql.NullString
	PaypalUpdateTime *time.Time
	Links            sql.NullString
	Data             sql.NullString
}
