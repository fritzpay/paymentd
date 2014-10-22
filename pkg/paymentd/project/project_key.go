package project

import (
	"encoding/hex"
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

// IsValid returns true if the project key is considered valid
func (p Projectkey) IsValid() bool {
	return p.Key != "" && p.Active
}

// SecretBytes returns the binary representation of the shared secret
func (p Projectkey) SecretBytes() ([]byte, error) {
	return hex.DecodeString(p.Secret)
}
