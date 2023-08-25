package models

import (
	"time"
)

// User represents a user of the services.
type User struct {
	ID        int
	Segments  []Segment
	CreatedAt time.Time
}
