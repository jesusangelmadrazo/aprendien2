package domain

import "time"

// ScheduleOverride pins a specific calendar date to something other than
// the instructor's recurring weekly pattern — a day off, or a one-off
// custom time range (e.g. "half day next Monday"). At most one per
// instructor+date; while it exists it fully replaces the recurring blocks
// for that date rather than adding to them.
type ScheduleOverride struct {
	ID           string
	InstructorID string
	Date         time.Time // truncated to the date, no time component
	IsDayOff     bool
	StartTime    string // "HH:MM", empty when IsDayOff
	EndTime      string
	CreatedAt    time.Time
}

type ScheduleOverrideRepository interface {
	// Upsert replaces any existing override for this instructor+date.
	Upsert(o *ScheduleOverride) error
	GetByInstructorAndDate(instructorID string, date time.Time) (*ScheduleOverride, error)
	ListByInstructorInRange(instructorID string, from, to time.Time) ([]*ScheduleOverride, error)
	ListUpcomingByInstructor(instructorID string, from time.Time) ([]*ScheduleOverride, error)
	Delete(instructorID string, date time.Time) error
}
