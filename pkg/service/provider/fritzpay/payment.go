package fritzpay

import (
	"time"
)

type Payment struct {
	ID        int64
	ProjectID int64
	PaymentID int64
	Created   time.Time
	MethodKey string
}
