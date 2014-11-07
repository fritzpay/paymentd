package project

import (
	"database/sql"
	"encoding/json"
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
	ID          int64 `json:",string"`
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

type Config struct {
	Timestamp          time.Time
	WebURL             sql.NullString
	CallbackURL        sql.NullString
	CallbackAPIVersion sql.NullString
	ProjectKey         sql.NullString
	ReturnURL          sql.NullString
}

type ConfigJSON struct {
	WebURL             *string
	CallbackURL        *string
	CallbackAPIVersion *string
	ProjectKey         *string
	ReturnURL          *string
}

// IsSet returns true if the config was set and stored
func (c Config) IsSet() bool {
	return !c.Timestamp.IsZero()
}

// HasValues returns true if the config has any values set
func (c Config) HasValues() bool {
	return c.WebURL.Valid || c.CallbackURL.Valid || c.CallbackAPIVersion.Valid || c.ProjectKey.Valid || c.ReturnURL.Valid
}

func (c *Config) UnmarshalJSON(p []byte) error {
	cfg := &ConfigJSON{}
	err := json.Unmarshal(p, cfg)
	if err != nil {
		return err
	}
	if cfg.WebURL != nil {
		c.WebURL.String, c.WebURL.Valid = *cfg.WebURL, true
	}
	if cfg.CallbackURL != nil {
		c.WebURL.String, c.WebURL.Valid = *cfg.WebURL, true
	}
	if cfg.CallbackAPIVersion != nil {
		c.CallbackAPIVersion.String, c.CallbackAPIVersion.Valid = *cfg.CallbackAPIVersion, true
	}
	if cfg.ProjectKey != nil {
		c.ProjectKey.String, c.ProjectKey.Valid = *cfg.ProjectKey, true
	}
	if cfg.ReturnURL != nil {
		c.ReturnURL.String, c.ReturnURL.Valid = *cfg.ReturnURL, true
	}
	return nil
}

func (c *Config) MarshalJSON() ([]byte, error) {
	cfg := &ConfigJSON{}
	if c.WebURL.Valid {
		cfg.WebURL = &c.WebURL.String
	}
	if c.CallbackURL.Valid {
		cfg.CallbackURL = &c.CallbackURL.String
	}
	if c.CallbackAPIVersion.Valid {
		cfg.CallbackAPIVersion = &c.CallbackAPIVersion.String
	}
	if c.ProjectKey.Valid {
		cfg.ProjectKey = &c.ProjectKey.String
	}
	if c.ReturnURL.Valid {
		cfg.ReturnURL = &c.ReturnURL.String
	}
	return json.Marshal(cfg)
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
