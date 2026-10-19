package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type AvailabilityRepo struct {
	pool *pgxpool.Pool
}

func NewAvailabilityRepo(pool *pgxpool.Pool) *AvailabilityRepo {
	return &AvailabilityRepo{pool: pool}
}

func (r *AvailabilityRepo) Create(b *domain.AvailabilityBlock) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO instructor_availability (instructor_id, day_of_week, start_time, end_time)
		VALUES ($1, $2, $3::time, $4::time)
		RETURNING id, created_at
	`, b.InstructorID, int(b.DayOfWeek), b.StartTime, b.EndTime).Scan(&b.ID, &b.CreatedAt)
}

func (r *AvailabilityRepo) ListByInstructor(instructorID string) ([]*domain.AvailabilityBlock, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT id, instructor_id, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at
		FROM instructor_availability WHERE instructor_id = $1 ORDER BY day_of_week, start_time
	`, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.AvailabilityBlock
	for rows.Next() {
		b := &domain.AvailabilityBlock{}
		var dayOfWeek int
		if err := rows.Scan(&b.ID, &b.InstructorID, &dayOfWeek, &b.StartTime, &b.EndTime, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.DayOfWeek = time.Weekday(dayOfWeek)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *AvailabilityRepo) Delete(id, instructorID string) error {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `DELETE FROM instructor_availability WHERE id = $1 AND instructor_id = $2`, id, instructorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
