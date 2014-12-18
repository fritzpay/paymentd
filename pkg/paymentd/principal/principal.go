package principal

import (
	"errors"
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

var (
	ErrInvalidStatus     = errors.New("invalid status")
	ErrPrincipalInactive = errors.New("principal is inactive")
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

// ValidStatus returns an ErrInvalidStatus if the set status is considered invalid
func (p Principal) ValidStatus() error {
	if p.Status != PrincipalStatusActive && p.Status != PrincipalStatusInactive && p.Status != PrincipalStatusDeleted {
		return ErrInvalidStatus
	}
	return nil
}

// Active will return an error if the status is not "active" or invalid
func (p Principal) Active() error {
	if err := p.ValidStatus(); err != nil {
		return err
	}
	if p.Status != PrincipalStatusActive {
		return ErrPrincipalInactive
	}
	return nil
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
