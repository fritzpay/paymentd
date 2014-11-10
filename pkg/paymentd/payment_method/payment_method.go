package payment_method

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"time"
)

type methodStatus string

// returns a valid paymentMethodStatus or error
func ParseMethodStatus(s string) (methodStatus, error) {
	if s == PaymentMethodStatusActive.String() {
		return PaymentMethodStatusActive, nil
	} else if s == PaymentMethodStatusInactive.String() {
		return PaymentMethodStatusInactive, nil
	} else {
		return methodStatus(""), errors.New("invalid")
	}
}

func (s methodStatus) String() string {
	if s == "" {
		return "invalid"
	}
	return string(s)
}

// Scan implements the (database/sql).Scanner
func (s *methodStatus) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		*s = methodStatus(string(v))
		return nil
	case string:
		*s = methodStatus(v)
		return nil
	default:
		return fmt.Errorf("error scanning into PaymentMethodStatus type. got invalid type %T", src)
	}
}

// Value implements the (database/sql/driver).Valuer so it can be used in SQL statements
// as a value
func (s methodStatus) Value() (driver.Value, error) {
	return string(s), nil
}

const (
	PaymentMethodStatusActive   methodStatus = "active"
	PaymentMethodStatusInactive methodStatus = "inactive"
)

// PaymentMethod represents a mode (method of payment)
//
// It is associated with a Provider and can be configured on a per-project base.
type Method struct {
	ID        int64 `json:",string"`
	ProjectID int64 `json:",string"`
	Provider  provider.Provider
	MethodKey string
	Created   time.Time
	CreatedBy string

	Status          methodStatus
	StatusChanged   time.Time
	StatusCreatedBy string

	Metadata map[string]string
}

// Active returns true if the payment method is considered active
func (m *Method) Active() bool {
	return m.Status == PaymentMethodStatusActive
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
