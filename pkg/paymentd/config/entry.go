package config

import (
	"errors"
	"time"
)

var (
	ErrEntryNotFound = errors.New("config entry not found")
)

// Entry represents a configuration entry
type Entry struct {
	Name       string
	lastChange time.Time
	Value      string
}

// Empty returns true if the entry is considered empty/not set
func (e Entry) Empty() bool {
	return e.Name == ""
}
