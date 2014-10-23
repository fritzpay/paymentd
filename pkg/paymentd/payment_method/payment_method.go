package payment_method

import (
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"time"
)

type PaymentMethodStatus string

func (s PaymentMethodStatus) String() string {
	if s == "" {
		return "invalid"
	}
	return string(s)
}

// Scan implements the (database/sql).Scanner
func (s *PaymentMethodStatus) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		*s = PaymentMethodStatus(string(v))
		return nil
	case string:
		*s = PaymentMethodStatus(v)
		return nil
	default:
		return fmt.Errorf("error scanning into PaymentMethodStatus type. got invalid type %T", src)
	}
}

const (
	PaymentMethodStatusActive   PaymentMethodStatus = "active"
	PaymentMethodStatusInactive PaymentMethodStatus = "inactive"
)

// PaymentMethod represents a mode (method of payment)
//
// It is associated with a Provider and can be configured on a per-project base.
type PaymentMethod struct {
	ID        int64
	ProjectID int64
	Provider  provider.Provider
	MethodKey string
	Created   time.Time
	CreatedBy string

	Status          PaymentMethodStatus
	StatusChanged   time.Time
	StatusCreatedBy string

	Metadata map[string]string
}

const (
	metadataTable        = "payment_method_metadata"
	metadataPrimaryField = "payment_method_id"
)

const MetadataModel metadataModel = 0

type metadataModel int

func (m metadataModel) Table() string {
	return metadataTable
}

func (m metadataModel) PrimaryField() string {
	return metadataPrimaryField
}
