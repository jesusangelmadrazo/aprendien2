package domain

import "time"

type LessonStatus string

const (
	LessonRequested        LessonStatus = "requested"
	LessonSearching        LessonStatus = "searching"
	LessonAccepted         LessonStatus = "accepted"
	LessonInstructorEnroute LessonStatus = "instructor_enroute"
	LessonInProgress       LessonStatus = "in_progress"
	LessonCompleted        LessonStatus = "completed"
	LessonCancelled        LessonStatus = "cancelled"
	LessonNoInstructorFound LessonStatus = "no_instructor_found"
)

type LessonRequest struct {
	ID                string
	StudentID         string
	InstructorID      *string
	Status            LessonStatus
	PickupLat         float64
	PickupLon         float64
	PickupAddressText string
	DurationMinutes   int
	EstimatedPriceUYU float64
	FinalPriceUYU     *float64
	// ScheduledAt is nil for an on-demand ("ahora") request handled by the
	// real-time Matcher, or set to a specific slot the student booked
	// directly from an instructor's published availability.
	ScheduledAt       *time.Time
	RequestedAt       time.Time
	AcceptedAt        *time.Time
	StartedAt         *time.Time
	CompletedAt       *time.Time
	CancelledAt       *time.Time
	CancelReason      string
}

type OfferResult string

const (
	OfferPending   OfferResult = "pending"
	OfferAccepted  OfferResult = "accepted"
	OfferRejected  OfferResult = "rejected"
	OfferTimedOut  OfferResult = "timed_out"
)

type LessonOffer struct {
	ID               string
	LessonRequestID  string
	InstructorID     string
	SequenceOrder    int
	Result           OfferResult
	OfferedAt        time.Time
	RespondedAt      *time.Time
}

type LessonRepository interface {
	Create(l *LessonRequest) error
	GetByID(id string) (*LessonRequest, error)
	AssignInstructor(lessonID, instructorID string) (bool, error)
	UpdateStatus(lessonID string, status LessonStatus) error
	Cancel(lessonID, reason string) error
	MarkNoInstructorFound(lessonID string) error
	ListByUser(userID string, role Role, status string) ([]*LessonRequest, error)
	ListScheduledByInstructor(instructorID string, from, to time.Time) ([]*LessonRequest, error)
	// CompletionStats returns how many lessons a student has completed and
	// the total minutes across them — the two metrics exam eligibility is
	// gated on (see ExamService).
	CompletionStats(studentID string) (lessonsCompleted int, minutesCompleted int, err error)

	CreateOffer(o *LessonOffer) error
	GetOffer(id string) (*LessonOffer, error)
	SetOfferResult(id string, result OfferResult) error
}
