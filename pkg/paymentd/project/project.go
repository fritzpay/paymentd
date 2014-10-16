package project

import (
	"time"
)

// Project represents a project
//
// A project is a resource of a principle.
// It has its payment methodes and can be used to separate
// different business units of one principle
type Project struct {
	ID          int64
	PrincipalID int64
	Name        string
	Created     time.Time
	CreatedBy   string

	Metadata map[string]string
}

// Empty returns true if the project is considered empty/uninitialized
func (p Project) Empty() bool {
	return p.ID == 0 && p.Name == ""
}
