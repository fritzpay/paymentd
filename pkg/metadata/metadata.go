package metadata

import (
	"time"
)

// MetadataModeler describes the construct for a concrete metadata model
type MetadataModeler interface {
	Table() string
	PrimaryField() string
}

// MetadataEntry represents an entry in metadata
type MetadataEntry struct {
	Name      string
	timestamp time.Time
	CreatedBy string
	Value     string
}

// IsEmpty returns true if this is an empty (non-existent) entry
func (e MetadataEntry) IsEmpty() bool {
	return e.Name == ""
}

// Metadata is a collection of metadata entries
type Metadata map[string]MetadataEntry

// MetadataFromValues creates metadata from a key-value map
func MetadataFromValues(values map[string]string, createdBy string) Metadata {
	metadata := make(map[string]MetadataEntry)
	for k, v := range values {
		metadata[k] = MetadataEntry{
			Name:      k,
			CreatedBy: createdBy,
			Value:     v,
		}
	}
	return metadata
}

// Values returns a flattened metadata map as key-values
func (m Metadata) Values() map[string]string {
	values := make(map[string]string)
	for _, e := range m {
		values[e.Name] = e.Value
	}
	return values
}
