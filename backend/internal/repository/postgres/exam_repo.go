package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type ExamRepo struct {
	pool *pgxpool.Pool
}

func NewExamRepo(pool *pgxpool.Pool) *ExamRepo {
	return &ExamRepo{pool: pool}
}

func (r *ExamRepo) Create(e *domain.Exam) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO exams (student_id, type, status, photo_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, submitted_at, created_at, updated_at
	`, e.StudentID, e.Type, e.Status, e.PhotoURL).Scan(&e.ID, &e.SubmittedAt, &e.CreatedAt, &e.UpdatedAt)
}

func (r *ExamRepo) GetLatestByStudentAndType(studentID string, examType domain.ExamType) (*domain.Exam, error) {
	ctx := context.Background()
	e := &domain.Exam{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, student_id, type, status, photo_url, COALESCE(rejection_reason, ''), reviewed_by,
		       submitted_at, reviewed_at, created_at, updated_at
		FROM exams WHERE student_id = $1 AND type = $2
		ORDER BY submitted_at DESC LIMIT 1
	`, studentID, examType).Scan(
		&e.ID, &e.StudentID, &e.Type, &e.Status, &e.PhotoURL, &e.RejectionReason, &e.ReviewedBy,
		&e.SubmittedAt, &e.ReviewedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *ExamRepo) GetByID(id string) (*domain.Exam, error) {
	ctx := context.Background()
	e := &domain.Exam{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, student_id, type, status, photo_url, COALESCE(rejection_reason, ''), reviewed_by,
		       submitted_at, reviewed_at, created_at, updated_at
		FROM exams WHERE id = $1
	`, id).Scan(
		&e.ID, &e.StudentID, &e.Type, &e.Status, &e.PhotoURL, &e.RejectionReason, &e.ReviewedBy,
		&e.SubmittedAt, &e.ReviewedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *ExamRepo) ListPendingReview() ([]*domain.Exam, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT id, student_id, type, status, photo_url, COALESCE(rejection_reason, ''), reviewed_by,
		       submitted_at, reviewed_at, created_at, updated_at
		FROM exams WHERE status = 'pending_review' ORDER BY submitted_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Exam
	for rows.Next() {
		e := &domain.Exam{}
		if err := rows.Scan(
			&e.ID, &e.StudentID, &e.Type, &e.Status, &e.PhotoURL, &e.RejectionReason, &e.ReviewedBy,
			&e.SubmittedAt, &e.ReviewedAt, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ExamRepo) Review(id string, status domain.ExamStatus, reviewedBy, rejectionReason string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		UPDATE exams SET status = $1, reviewed_by = $2, rejection_reason = $3, reviewed_at = now(), updated_at = now()
		WHERE id = $4
	`, status, reviewedBy, rejectionReason, id)
	return err
}
