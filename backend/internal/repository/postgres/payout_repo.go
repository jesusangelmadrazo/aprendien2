package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type PayoutMethodRepo struct {
	pool *pgxpool.Pool
}

func NewPayoutMethodRepo(pool *pgxpool.Pool) *PayoutMethodRepo {
	return &PayoutMethodRepo{pool: pool}
}

func (r *PayoutMethodRepo) Upsert(m *domain.PayoutMethod) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO instructor_payout_methods (instructor_id, bank_name, account_type, account_number, account_holder_name, account_holder_document)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (instructor_id) DO UPDATE SET
			bank_name = EXCLUDED.bank_name,
			account_type = EXCLUDED.account_type,
			account_number = EXCLUDED.account_number,
			account_holder_name = EXCLUDED.account_holder_name,
			account_holder_document = EXCLUDED.account_holder_document,
			updated_at = now()
		RETURNING updated_at
	`, m.InstructorID, m.BankName, m.AccountType, m.AccountNumber, m.AccountHolderName, m.AccountHolderDocument).
		Scan(&m.UpdatedAt)
}

func (r *PayoutMethodRepo) GetByInstructor(instructorID string) (*domain.PayoutMethod, error) {
	ctx := context.Background()
	m := &domain.PayoutMethod{}
	err := r.pool.QueryRow(ctx, `
		SELECT instructor_id, bank_name, account_type, account_number, account_holder_name, account_holder_document, updated_at
		FROM instructor_payout_methods WHERE instructor_id = $1
	`, instructorID).Scan(
		&m.InstructorID, &m.BankName, &m.AccountType, &m.AccountNumber, &m.AccountHolderName, &m.AccountHolderDocument, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

type PayoutRepo struct {
	pool *pgxpool.Pool
}

func NewPayoutRepo(pool *pgxpool.Pool) *PayoutRepo {
	return &PayoutRepo{pool: pool}
}

func (r *PayoutRepo) Create(p *domain.Payout) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO payouts (instructor_id, amount_uyu, period_start, period_end, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, p.InstructorID, p.AmountUYU, p.PeriodStart, p.PeriodEnd, p.CreatedBy).Scan(&p.ID, &p.CreatedAt)
}

func (r *PayoutRepo) ListByInstructor(instructorID string) ([]*domain.Payout, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT id, instructor_id, amount_uyu, period_start, period_end, created_by, created_at
		FROM payouts WHERE instructor_id = $1 ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Payout
	for rows.Next() {
		p := &domain.Payout{}
		if err := rows.Scan(&p.ID, &p.InstructorID, &p.AmountUYU, &p.PeriodStart, &p.PeriodEnd, &p.CreatedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
