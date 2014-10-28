package project

import (
	"database/sql"
	"time"
)

const (
	metadataTable        = "project_metadata"
	metadataPrimaryField = "project_id"
)

// Project represents a project
//
// A project is a resource of a principle.
// It has its payment methodes and can be used to separate
// different business units of one principle
type Project struct {
	ID          int64
	PrincipalID int64 `json:",string"`
	Name        string
	Created     time.Time
	CreatedBy   string

	Config Config

	Metadata map[string]string
}

// Empty returns true if the project is considered empty/uninitialized
func (p *Project) Empty() bool {
	return p.ID == 0 && p.Name == ""
}

// Validates if the obligatory fields are set
func (p *Project) IsValid() bool {
	if len(p.Name) < 1 || len(p.CreatedBy) < 1 {
		return false
	}
	return true
}

type Config struct {
	Timestamp          time.Time
	WebURL             sql.NullString
	CallbackURL        sql.NullString
	CallbackAPIVersion sql.NullString
	ProjectKey         sql.NullString
	ReturnURL          sql.NullString
}

// IsSet returns true if the config was set
func (c Config) IsSet() bool {
	return !c.Timestamp.IsZero()
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
