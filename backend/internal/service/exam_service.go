package service

import (
	"errors"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
)

var (
	ErrInvalidExamType      = errors.New("invalid exam type")
	ErrExamLocked           = errors.New("exam is not unlocked yet")
	ErrExamAlreadySubmitted = errors.New("exam already submitted and awaiting review or approved")
)

// ExamTypeStatus is whether one exam type is unlocked for a student yet
// (computed from their lesson history, never stored) plus their latest
// submission for it, if any.
type ExamTypeStatus struct {
	Type       domain.ExamType `json:"type"`
	Unlocked   bool            `json:"unlocked"`
	Progress   int             `json:"progress"`
	Threshold  int             `json:"threshold"`
	Submission *domain.Exam    `json:"submission"`
}

type ExamStatusOverview struct {
	Theory   ExamTypeStatus `json:"theory"`
	Practice ExamTypeStatus `json:"practice"`
}

type ExamService struct {
	exams                    domain.ExamRepository
	lessons                  domain.LessonRepository
	theoryThresholdLessons   int
	practiceThresholdMinutes int
}

func NewExamService(exams domain.ExamRepository, lessons domain.LessonRepository, theoryThresholdLessons, practiceThresholdMinutes int) *ExamService {
	return &ExamService{
		exams:                    exams,
		lessons:                  lessons,
		theoryThresholdLessons:   theoryThresholdLessons,
		practiceThresholdMinutes: practiceThresholdMinutes,
	}
}

func (s *ExamService) GetStatus(studentID string) (*ExamStatusOverview, error) {
	lessonsCompleted, minutesCompleted, err := s.lessons.CompletionStats(studentID)
	if err != nil {
		return nil, err
	}

	theorySubmission, err := s.latestOrNil(studentID, domain.ExamTheory)
	if err != nil {
		return nil, err
	}
	practiceSubmission, err := s.latestOrNil(studentID, domain.ExamPractice)
	if err != nil {
		return nil, err
	}

	return &ExamStatusOverview{
		Theory: ExamTypeStatus{
			Type:       domain.ExamTheory,
			Unlocked:   lessonsCompleted >= s.theoryThresholdLessons,
			Progress:   lessonsCompleted,
			Threshold:  s.theoryThresholdLessons,
			Submission: theorySubmission,
		},
		Practice: ExamTypeStatus{
			Type:       domain.ExamPractice,
			Unlocked:   minutesCompleted >= s.practiceThresholdMinutes,
			Progress:   minutesCompleted,
			Threshold:  s.practiceThresholdMinutes,
			Submission: practiceSubmission,
		},
	}, nil
}

func (s *ExamService) latestOrNil(studentID string, examType domain.ExamType) (*domain.Exam, error) {
	exam, err := s.exams.GetLatestByStudentAndType(studentID, examType)
	if errors.Is(err, postgres.ErrNotFound) {
		return nil, nil
	}
	return exam, err
}

// Submit records a student's proof of passing an exam. Resubmission is only
// allowed after a rejection — an approved or still-pending submission blocks
// a new one, since the whole point is that this can't be gamed by retrying
// until someone approves it.
func (s *ExamService) Submit(studentID string, examType domain.ExamType, photoURL string) (*domain.Exam, error) {
	status, err := s.GetStatus(studentID)
	if err != nil {
		return nil, err
	}

	var typeStatus ExamTypeStatus
	switch examType {
	case domain.ExamTheory:
		typeStatus = status.Theory
	case domain.ExamPractice:
		typeStatus = status.Practice
	default:
		return nil, ErrInvalidExamType
	}

	if !typeStatus.Unlocked {
		return nil, ErrExamLocked
	}
	if typeStatus.Submission != nil && typeStatus.Submission.Status != domain.ExamRejected {
		return nil, ErrExamAlreadySubmitted
	}

	exam := &domain.Exam{StudentID: studentID, Type: examType, Status: domain.ExamPendingReview, PhotoURL: photoURL}
	if err := s.exams.Create(exam); err != nil {
		return nil, err
	}
	return exam, nil
}

func (s *ExamService) ListPendingReview() ([]*domain.Exam, error) {
	return s.exams.ListPendingReview()
}

func (s *ExamService) Review(examID, adminID string, approve bool, reason string) error {
	status := domain.ExamRejected
	if approve {
		status = domain.ExamApproved
	}
	return s.exams.Review(examID, status, adminID, reason)
}
