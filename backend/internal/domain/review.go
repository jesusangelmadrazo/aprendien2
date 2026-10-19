package domain

import "time"

type Review struct {
	ID              string
	LessonRequestID string
	ReviewerID      string
	RevieweeID      string
	Rating          int
	Comment         string
	CreatedAt       time.Time
}

type ReviewRepository interface {
	Create(r *Review) error
	ListByInstructor(instructorID string) ([]*Review, error)
}
