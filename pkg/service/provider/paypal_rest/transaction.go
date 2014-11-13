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

// Transaction represents a transaction on a paypal payment
//
// It can be one of the following:
//
//   - A representation of a request.
//   - A representation of a response.
//   - A representation of a local change to a transaction.
//
// It also keeps the state of the paypal payment, i.e. the most recent transaction
// will denote what the state of the payment is.
type Transaction struct {
	ProjectID        int64
	PaymentID        int64
	Timestamp        time.Time
	Type             string
	Intent           sql.NullString
	PaypalID         sql.NullString
	PayerID          sql.NullString
	PaypalCreateTime *time.Time
	PaypalState      sql.NullString
	PaypalUpdateTime *time.Time
	Links            []byte
	Data             []byte
}

func (t *Transaction) SetIntent(intent string) {
	t.Intent.String, t.Intent.Valid = intent, true
}

func (t *Transaction) SetPaypalID(id string) {
	t.PaypalID.String, t.PaypalID.Valid = id, true
}

func (t *Transaction) SetPayerID(id string) {
	t.PayerID.String, t.PayerID.Valid = id, true
}

func (t *Transaction) SetState(state string) {
	t.PaypalState.String, t.PaypalState.Valid = state, true
}
