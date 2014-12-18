package principal

import (
	"time"
)

const (
	metadataTable        = "principal_metadata"
	metadataPrimaryField = "principal_id"
)

const (
	PrincipalStatusActive   = "active"
	PrincipalStatusInactive = "inactive"
	PrincipalStatusDeleted  = "deleted"
)

// Principal represents a principal
//
// A principal is a resource under which projects are organized
type Principal struct {
	ID        int64 `json:",string"`
	Created   time.Time
	CreatedBy string
	Name      string

	Status string

	Metadata map[string]string
}

// Empty returns true if the principal is considered empty/uninitialized
func (p Principal) Empty() bool {
	return p.ID == 0 && p.Name == ""
}

// representation of the metadata schema structure
const MetadataModel metadataModel = 0

// pattern for nicer package usage
// this prevents the initialistation of a struct object{}
// instead devs can just take the MetadataModel const
type metadataModel int

func (m metadataModel) Table() string {
	return metadataTable
}

func (m metadataModel) PrimaryField() string {
	return metadataPrimaryField
}
