package payment

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/decimal"
	"time"
)

type PaymentTransactionStatus string

// Scan implements the Scanner interface for sql
func (s *PaymentTransactionStatus) Scan(v interface{}) error {
	switch src := v.(type) {
	case []byte:
		*s = PaymentTransactionStatus(string(src))
		return nil
	}
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into %T", v, s)
	}
	*s = PaymentTransactionStatus(str)
	return nil
}

// Value implements the Valuer interface for sql
func (s PaymentTransactionStatus) Value() (driver.Value, error) {
	return driver.Value(s.String()), nil
}

func (s PaymentTransactionStatus) String() string {
	return string(s)
}

const (
	PaymentStatusOpen           PaymentTransactionStatus = "open"
	PaymentStatusPending                                 = "pending"
	PaymentStatusPaid                                    = "paid"
	PaymentStatusSettled                                 = "settled"
	PaymentStatusAuthorized                              = "authorized"
	PaymentStatusError                                   = "error"
	PaymentStatusCancelled                               = "cancelled"
	PaymentStatusFailed                                  = "failed"
	PaymentStatusChargeback                              = "chargeback"
	PaymentStatusRefunded                                = "refunded"
	PaymentStatusRefundReversed                          = "refund-reversed"
)

// PaymentTransaction represents a transaction on a payment
//
// A transaction is any event/status change on a payment
//
// The transactions represents the ledger on a payment
type PaymentTransaction struct {
	Payment *Payment

	Timestamp time.Time
	Amount    int64
	Subunits  int8
	Currency  string
	Status    PaymentTransactionStatus
	Comment   sql.NullString
}

// Balance represents a balance which totals the ledger by currency
type Balance map[string]*decimal.Decimal
