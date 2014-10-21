package project

import (
	"time"
)

// ProjectKey represents a project key
type Projectkey struct {
	Key         string
	Timestamp   time.Time
	Project     Project
	CreatedBy   string
	Secret      string
	secretBytes []byte
	Active      bool
}
