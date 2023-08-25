package models

import (
	"time"
)

// Segment represents a segment that users can belong to.
type Segment struct {
	ID        int
	Slug      string
	AutoAdd   bool
	AutoPct   int
	CreatedAt time.Time
}
