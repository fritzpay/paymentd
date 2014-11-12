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
	Links            []byte
	Data             []byte
}

func (t *Transaction) SetPaypalID(id string) {
	t.PaypalID.String, t.PaypalID.Valid = id, true
}

func (t *Transaction) SetState(state string) {
	t.PaypalState.String, t.PaypalState.Valid = state, true
}
