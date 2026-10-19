package domain

import "time"

type PaymentStatus string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentProcessing PaymentStatus = "processing"
	PaymentApproved   PaymentStatus = "approved"
	PaymentRejected   PaymentStatus = "rejected"
	PaymentRefunded   PaymentStatus = "refunded"
)

// Payment records what a student was charged for a lesson and how it splits
// between the instructor and the platform. InstructorPriceUYU is always
// exactly what the instructor quoted; PlatformFeeUYU is the service fee
// added on top, so AmountUYU == InstructorPriceUYU + PlatformFeeUYU.
// PayoutID stays nil until this payment's instructor share is included in a
// settled weekly Payout.
type Payment struct {
	ID                 string
	LessonRequestID    string
	StudentID          string
	InstructorID       *string
	AmountUYU          float64
	InstructorPriceUYU float64
	PlatformFeeUYU     float64
	ProviderRef        string
	ProviderPaymentID  string
	Status             PaymentStatus
	PayoutID           *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type PaymentRepository interface {
	Create(p *Payment) error
	GetByLessonID(lessonID string) (*Payment, error)
	UpdateStatus(id string, status PaymentStatus, providerPaymentID string) error
	GetByProviderRef(providerRef string) (*Payment, error)
	ListByStudent(studentID string) ([]*Payment, error)

	// ListPendingByInstructor returns approved payments not yet in a payout.
	ListPendingByInstructor(instructorID string) ([]*Payment, error)
	// ListPendingBalances sums pending amounts per instructor, for the
	// admin's "who's owed money" queue.
	ListPendingBalances() ([]InstructorBalance, error)
	// MarkPaidForInstructor stamps payoutID onto every pending payment for
	// that instructor — called right after creating the Payout row.
	MarkPaidForInstructor(instructorID, payoutID string) error
}
