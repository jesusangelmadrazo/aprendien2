package domain

import "time"

type ExamType string

const (
	ExamTheory   ExamType = "theory"
	ExamPractice ExamType = "practice"
)

type ExamStatus string

const (
	ExamPendingReview ExamStatus = "pending_review"
	ExamApproved      ExamStatus = "approved"
	ExamRejected      ExamStatus = "rejected"
)

// Exam is a student's submission of proof (a photo of the official result)
// that they passed the theory or practice exam. It only exists in the DB
// once submitted — eligibility ("unlocked") before that point is computed
// on the fly from the student's lesson history, not stored.
type Exam struct {
	ID              string
	StudentID       string
	Type            ExamType
	Status          ExamStatus
	PhotoURL        string
	RejectionReason string
	ReviewedBy      *string
	SubmittedAt     time.Time
	ReviewedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ExamRepository interface {
	Create(e *Exam) error
	GetLatestByStudentAndType(studentID string, examType ExamType) (*Exam, error)
	GetByID(id string) (*Exam, error)
	ListPendingReview() ([]*Exam, error)
	Review(id string, status ExamStatus, reviewedBy, rejectionReason string) error
}
