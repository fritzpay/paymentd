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
