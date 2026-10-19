package domain

import "time"

type DocumentType string

const (
	DocLicense             DocumentType = "license"             // libreta de conducir
	DocIdentity            DocumentType = "identity"             // cédula de identidad de la persona
	DocVehicleRegistration DocumentType = "vehicle_registration" // propiedad del auto
	DocInsurance           DocumentType = "insurance"
)

type DocumentStatus string

const (
	DocPendingReview DocumentStatus = "pending_review"
	DocApproved      DocumentStatus = "approved"
	DocRejected      DocumentStatus = "rejected"
)

// VerificationStatus is the instructor's overall onboarding gate: only
// "approved" once every required document is approved and unexpired. It's
// a cached field on InstructorProfile, recomputed by DocumentService
// whenever an admin reviews a document.
type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationApproved VerificationStatus = "approved"
)

// InstructorDocument is one submitted piece of onboarding paperwork —
// license, vehicle registration, or insurance — with the date it's valid
// through (nil if it doesn't expire) and its review state.
type InstructorDocument struct {
	ID              string
	InstructorID    string
	Type            DocumentType
	PhotoURL        string
	ExpiryDate      *time.Time
	Status          DocumentStatus
	RejectionReason string
	ReviewedBy      *string
	SubmittedAt     time.Time
	ReviewedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InstructorDocumentRepository interface {
	Create(d *InstructorDocument) error
	GetByID(id string) (*InstructorDocument, error)
	GetLatestByInstructorAndType(instructorID string, docType DocumentType) (*InstructorDocument, error)
	ListByInstructor(instructorID string) ([]*InstructorDocument, error)
	ListPendingReview() ([]*InstructorDocument, error)
	Review(id string, status DocumentStatus, reviewedBy, rejectionReason string) error
}
