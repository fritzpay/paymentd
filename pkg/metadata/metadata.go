package metadata

import (
	"time"
)

// MetadataModel describes the concrete metadata model
type MetadataModel interface {
	Schema() string
	PrimaryField() string
}

// MetadataEntry represents an entry in metadata
type MetadataEntry struct {
	Name      string
	Timestamp time.Time
	CreatedBy string
	Value     string
}

// IsEmpty returns true if this is an empty entry
func (e MetadataEntry) IsEmpty() bool {
	return e.Name == ""
}

// Metadata is a collection of metadata entries
type Metadata map[string]MetadataEntry
