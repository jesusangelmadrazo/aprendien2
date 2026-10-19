package domain

import "time"

// AvailabilityBlock is a recurring weekly window during which an instructor
// takes bookings (e.g. "Mondays 09:00-12:00"). StartTime/EndTime are "HH:MM"
// strings rather than time.Time — there's no date to anchor them to, they
// repeat every week.
type AvailabilityBlock struct {
	ID           string
	InstructorID string
	DayOfWeek    time.Weekday // 0 = Sunday .. 6 = Saturday
	StartTime    string
	EndTime      string
	CreatedAt    time.Time
}

type AvailabilityRepository interface {
	Create(b *AvailabilityBlock) error
	ListByInstructor(instructorID string) ([]*AvailabilityBlock, error)
	Delete(id, instructorID string) error
}
