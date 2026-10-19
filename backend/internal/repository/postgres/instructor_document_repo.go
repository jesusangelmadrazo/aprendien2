package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type InstructorDocumentRepo struct {
	pool *pgxpool.Pool
}

func NewInstructorDocumentRepo(pool *pgxpool.Pool) *InstructorDocumentRepo {
	return &InstructorDocumentRepo{pool: pool}
}

const documentColumns = `
	id, instructor_id, doc_type, photo_url, expiry_date, status,
	COALESCE(rejection_reason, ''), reviewed_by, submitted_at, reviewed_at, created_at, updated_at
`

func scanDocument(row interface{ Scan(...any) error }, d *domain.InstructorDocument) error {
	return row.Scan(
		&d.ID, &d.InstructorID, &d.Type, &d.PhotoURL, &d.ExpiryDate, &d.Status,
		&d.RejectionReason, &d.ReviewedBy, &d.SubmittedAt, &d.ReviewedAt, &d.CreatedAt, &d.UpdatedAt,
	)
}

func (r *InstructorDocumentRepo) Create(d *domain.InstructorDocument) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO instructor_documents (instructor_id, doc_type, photo_url, expiry_date, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, submitted_at, created_at, updated_at
	`, d.InstructorID, d.Type, d.PhotoURL, d.ExpiryDate, d.Status).Scan(&d.ID, &d.SubmittedAt, &d.CreatedAt, &d.UpdatedAt)
}

func (r *InstructorDocumentRepo) GetByID(id string) (*domain.InstructorDocument, error) {
	ctx := context.Background()
	d := &domain.InstructorDocument{}
	err := scanDocument(r.pool.QueryRow(ctx, `SELECT `+documentColumns+` FROM instructor_documents WHERE id = $1`, id), d)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *InstructorDocumentRepo) GetLatestByInstructorAndType(instructorID string, docType domain.DocumentType) (*domain.InstructorDocument, error) {
	ctx := context.Background()
	d := &domain.InstructorDocument{}
	err := scanDocument(r.pool.QueryRow(ctx, `
		SELECT `+documentColumns+` FROM instructor_documents
		WHERE instructor_id = $1 AND doc_type = $2
		ORDER BY submitted_at DESC LIMIT 1
	`, instructorID, docType), d)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *InstructorDocumentRepo) ListByInstructor(instructorID string) ([]*domain.InstructorDocument, error) {
	return r.queryList(`SELECT `+documentColumns+` FROM instructor_documents WHERE instructor_id = $1 ORDER BY submitted_at DESC`, instructorID)
}

func (r *InstructorDocumentRepo) ListPendingReview() ([]*domain.InstructorDocument, error) {
	return r.queryList(`SELECT ` + documentColumns + ` FROM instructor_documents WHERE status = 'pending_review' ORDER BY submitted_at ASC`)
}

func (r *InstructorDocumentRepo) queryList(query string, args ...interface{}) ([]*domain.InstructorDocument, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.InstructorDocument
	for rows.Next() {
		d := &domain.InstructorDocument{}
		if err := scanDocument(rows, d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *InstructorDocumentRepo) Review(id string, status domain.DocumentStatus, reviewedBy, rejectionReason string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		UPDATE instructor_documents SET status = $1, reviewed_by = $2, rejection_reason = $3, reviewed_at = now(), updated_at = now()
		WHERE id = $4
	`, status, reviewedBy, rejectionReason, id)
	return err
}
