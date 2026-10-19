package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

const paymentColumns = `
	id, lesson_request_id, student_id, instructor_id, amount_uyu, instructor_price_uyu, platform_fee_uyu,
	COALESCE(provider_ref,''), COALESCE(provider_payment_id,''), status, payout_id, created_at, updated_at
`

func scanPayment(row interface{ Scan(...any) error }, p *domain.Payment) error {
	return row.Scan(
		&p.ID, &p.LessonRequestID, &p.StudentID, &p.InstructorID, &p.AmountUYU, &p.InstructorPriceUYU, &p.PlatformFeeUYU,
		&p.ProviderRef, &p.ProviderPaymentID, &p.Status, &p.PayoutID, &p.CreatedAt, &p.UpdatedAt,
	)
}

func (r *PaymentRepo) Create(p *domain.Payment) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO payments (lesson_request_id, student_id, instructor_id, amount_uyu, instructor_price_uyu, platform_fee_uyu, provider_ref, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, p.LessonRequestID, p.StudentID, p.InstructorID, p.AmountUYU, p.InstructorPriceUYU, p.PlatformFeeUYU, p.ProviderRef, p.Status).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PaymentRepo) GetByLessonID(lessonID string) (*domain.Payment, error) {
	ctx := context.Background()
	p := &domain.Payment{}
	err := scanPayment(r.pool.QueryRow(ctx, `SELECT `+paymentColumns+` FROM payments WHERE lesson_request_id = $1`, lessonID), p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepo) GetByProviderRef(providerRef string) (*domain.Payment, error) {
	ctx := context.Background()
	p := &domain.Payment{}
	err := scanPayment(r.pool.QueryRow(ctx, `SELECT `+paymentColumns+` FROM payments WHERE provider_ref = $1`, providerRef), p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepo) ListByStudent(studentID string) ([]*domain.Payment, error) {
	return r.queryList(`SELECT `+paymentColumns+` FROM payments WHERE student_id = $1 ORDER BY created_at DESC`, studentID)
}

func (r *PaymentRepo) ListPendingByInstructor(instructorID string) ([]*domain.Payment, error) {
	return r.queryList(`
		SELECT `+paymentColumns+`
		FROM payments WHERE instructor_id = $1 AND status = 'approved' AND payout_id IS NULL
		ORDER BY created_at ASC
	`, instructorID)
}

func (r *PaymentRepo) queryList(query string, args ...interface{}) ([]*domain.Payment, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := scanPayment(rows, p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PaymentRepo) ListPendingBalances() ([]domain.InstructorBalance, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT p.instructor_id, u.full_name, SUM(p.instructor_price_uyu)
		FROM payments p
		JOIN users u ON u.id = p.instructor_id
		WHERE p.status = 'approved' AND p.payout_id IS NULL AND p.instructor_id IS NOT NULL
		GROUP BY p.instructor_id, u.full_name
		ORDER BY u.full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.InstructorBalance
	for rows.Next() {
		var b domain.InstructorBalance
		if err := rows.Scan(&b.InstructorID, &b.InstructorName, &b.AmountUYU); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *PaymentRepo) MarkPaidForInstructor(instructorID, payoutID string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		UPDATE payments SET payout_id = $1, updated_at = now()
		WHERE instructor_id = $2 AND status = 'approved' AND payout_id IS NULL
	`, payoutID, instructorID)
	return err
}

func (r *PaymentRepo) UpdateStatus(id string, status domain.PaymentStatus, providerPaymentID string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		UPDATE payments SET status = $1, provider_payment_id = $2, updated_at = now() WHERE id = $3
	`, status, providerPaymentID, id)
	return err
}
