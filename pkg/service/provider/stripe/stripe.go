package stripe

import (
	"time"
)

type Config struct {
	ProjectID int64
	MethodKey string
	Created   time.Time
	CreatedBy string

	SecretKey string
	PublicKey string
}
