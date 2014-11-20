package paypal_rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

const (
	TransactionTypeCreatePayment          = "createPayment"
	TransactionTypeCreatePaymentResponse  = "createPaymentResponse"
	TransactionTypeError                  = "error"
	TransactionTypeCancelled              = "cancelled"
	TransactionTypeExecutePayment         = "executePayment"
	TransactionTypeExecutePaymentResponse = "executePaymentResponse"
)

var (
	ErrNoLinks = errors.New("no links")
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
	Nonce            sql.NullString
	Intent           sql.NullString
	PaypalID         sql.NullString
	PayerID          sql.NullString
	PaypalCreateTime *time.Time
	PaypalState      sql.NullString
	PaypalUpdateTime *time.Time
	Links            []byte
	Data             []byte
}

func (t *Transaction) SetNonce(nonce string) {
	t.Nonce.String, t.Nonce.Valid = nonce, true
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

func (t *Transaction) PayPalLinks() (map[string]*PayPalLink, error) {
	if t.Links == nil || len(t.Links) == 0 {
		return nil, ErrNoLinks
	}
	links := make([]*PayPalLink, 0, 10)
	err := json.Unmarshal(t.Links, &links)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]*PayPalLink)
	for _, l := range links {
		ret[l.Rel] = l
	}
	return ret, nil
}

func NewPayPalPaymentTransaction(paypalP *PaypalPayment) (*Transaction, error) {
	var err error
	paypalTx := &Transaction{
		Timestamp: time.Now(),
	}
	if paypalP.Intent != "" {
		paypalTx.SetIntent(paypalP.Intent)
	}
	if paypalP.ID != "" {
		paypalTx.SetPaypalID(paypalP.ID)
	}
	if paypalP.State != "" {
		paypalTx.SetState(paypalP.State)
	}
	if paypalP.CreateTime != "" {
		var t time.Time
		t, err = time.Parse(time.RFC3339, paypalP.CreateTime)
		if err == nil {
			paypalTx.PaypalCreateTime = &t
		}
	}
	if paypalP.UpdateTime != "" {
		var t time.Time
		t, err = time.Parse(time.RFC3339, paypalP.UpdateTime)
		if err == nil {
			paypalTx.PaypalUpdateTime = &t
		}
	}
	paypalTx.Links, err = json.Marshal(paypalP.Links)
	if err != nil {
		return nil, err
	}
	paypalTx.Data, err = json.Marshal(paypalP)
	if err != nil {
		return nil, err
	}
	return paypalTx, err
}
