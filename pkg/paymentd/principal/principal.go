package principal

import (
	"time"
)

// Principal represents a principal
//
// A principal is a resource under which projects are organized
type Principal struct {
	ID        int64
	Created   time.Time
	CreatedBy string
	Name      string
}
