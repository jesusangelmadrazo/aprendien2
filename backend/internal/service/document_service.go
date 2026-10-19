package service

import (
	"errors"
	"time"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
)

var (
	ErrDocumentAlreadySubmitted = errors.New("document already submitted and awaiting review or approved")
)

// requiredDocTypes is every document an instructor must have approved
// (and, where applicable, not expired) before they're considered verified.
var requiredDocTypes = []domain.DocumentType{domain.DocLicense, domain.DocIdentity, domain.DocVehicleRegistration, domain.DocInsurance}

type DocumentStatusOverview struct {
	Verified  bool                                             `json:"verified"`
	Documents map[domain.DocumentType]*domain.InstructorDocument `json:"documents"`
}

// DocumentService is the instructor onboarding gate: license, vehicle
// registration, and insurance each need an admin-approved, unexpired
// submission before the instructor counts as verified — see
// InstructorRepository.ListAll (only lists verified instructors) and
// InstructorService.SetStatus (blocks going online otherwise).
type DocumentService struct {
	documents   domain.InstructorDocumentRepository
	instructors domain.InstructorRepository
}

func NewDocumentService(documents domain.InstructorDocumentRepository, instructors domain.InstructorRepository) *DocumentService {
	return &DocumentService{documents: documents, instructors: instructors}
}

func (s *DocumentService) GetStatus(instructorID string) (*DocumentStatusOverview, error) {
	overview := &DocumentStatusOverview{Documents: make(map[domain.DocumentType]*domain.InstructorDocument)}
	verified := true
	for _, dt := range requiredDocTypes {
		doc, err := s.latestOrNil(instructorID, dt)
		if err != nil {
			return nil, err
		}
		overview.Documents[dt] = doc
		if !isValidApproval(doc) {
			verified = false
		}
	}
	overview.Verified = verified
	return overview, nil
}

func (s *DocumentService) latestOrNil(instructorID string, docType domain.DocumentType) (*domain.InstructorDocument, error) {
	doc, err := s.documents.GetLatestByInstructorAndType(instructorID, docType)
	if errors.Is(err, postgres.ErrNotFound) {
		return nil, nil
	}
	return doc, err
}

func isValidApproval(doc *domain.InstructorDocument) bool {
	if doc == nil || doc.Status != domain.DocApproved {
		return false
	}
	if doc.ExpiryDate != nil && doc.ExpiryDate.Before(time.Now()) {
		return false
	}
	return true
}

// Submit records one document. Resubmission is only allowed after a
// rejection — a still-pending or already-approved submission blocks a new
// one, same rule as exams and for the same reason: this paperwork is what
// the platform vouches for, it can't be gamed by spamming retries.
func (s *DocumentService) Submit(instructorID string, docType domain.DocumentType, photoURL string, expiryDate *time.Time) (*domain.InstructorDocument, error) {
	latest, err := s.latestOrNil(instructorID, docType)
	if err != nil {
		return nil, err
	}
	if latest != nil && latest.Status != domain.DocRejected {
		return nil, ErrDocumentAlreadySubmitted
	}

	doc := &domain.InstructorDocument{
		InstructorID: instructorID,
		Type:         docType,
		PhotoURL:     photoURL,
		ExpiryDate:   expiryDate,
		Status:       domain.DocPendingReview,
	}
	if err := s.documents.Create(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) ListPendingReview() ([]*domain.InstructorDocument, error) {
	return s.documents.ListPendingReview()
}

// Review approves or rejects a document, then recomputes and caches the
// instructor's overall verification status.
func (s *DocumentService) Review(docID, adminID string, approve bool, reason string) error {
	doc, err := s.documents.GetByID(docID)
	if err != nil {
		return err
	}

	status := domain.DocRejected
	if approve {
		status = domain.DocApproved
	}
	if err := s.documents.Review(docID, status, adminID, reason); err != nil {
		return err
	}

	overview, err := s.GetStatus(doc.InstructorID)
	if err != nil {
		return err
	}
	newStatus := domain.VerificationPending
	if overview.Verified {
		newStatus = domain.VerificationApproved
	}
	return s.instructors.UpdateVerificationStatus(doc.InstructorID, newStatus)
}
