package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type ScheduleOverrideRepo struct {
	pool *pgxpool.Pool
}

func NewScheduleOverrideRepo(pool *pgxpool.Pool) *ScheduleOverrideRepo {
	return &ScheduleOverrideRepo{pool: pool}
}

func (r *ScheduleOverrideRepo) Upsert(o *domain.ScheduleOverride) error {
	ctx := context.Background()
	var startTime, endTime interface{}
	if !o.IsDayOff {
		startTime, endTime = o.StartTime, o.EndTime
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO instructor_schedule_overrides (instructor_id, date, is_day_off, start_time, end_time)
		VALUES ($1, $2, $3, $4::time, $5::time)
		ON CONFLICT (instructor_id, date) DO UPDATE SET
			is_day_off = EXCLUDED.is_day_off,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time
		RETURNING id, created_at
	`, o.InstructorID, o.Date, o.IsDayOff, startTime, endTime).Scan(&o.ID, &o.CreatedAt)
}

func (r *ScheduleOverrideRepo) GetByInstructorAndDate(instructorID string, date time.Time) (*domain.ScheduleOverride, error) {
	ctx := context.Background()
	o := &domain.ScheduleOverride{}
	var startTime, endTime *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, instructor_id, date, is_day_off, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at
		FROM instructor_schedule_overrides WHERE instructor_id = $1 AND date = $2
	`, instructorID, date).Scan(&o.ID, &o.InstructorID, &o.Date, &o.IsDayOff, &startTime, &endTime, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if startTime != nil {
		o.StartTime = *startTime
	}
	if endTime != nil {
		o.EndTime = *endTime
	}
	return o, nil
}

func (r *ScheduleOverrideRepo) ListByInstructorInRange(instructorID string, from, to time.Time) ([]*domain.ScheduleOverride, error) {
	return r.queryList(`
		SELECT id, instructor_id, date, is_day_off, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at
		FROM instructor_schedule_overrides WHERE instructor_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date ASC
	`, instructorID, from, to)
}

func (r *ScheduleOverrideRepo) ListUpcomingByInstructor(instructorID string, from time.Time) ([]*domain.ScheduleOverride, error) {
	return r.queryList(`
		SELECT id, instructor_id, date, is_day_off, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), created_at
		FROM instructor_schedule_overrides WHERE instructor_id = $1 AND date >= $2
		ORDER BY date ASC
	`, instructorID, from)
}

func (r *ScheduleOverrideRepo) queryList(query string, args ...interface{}) ([]*domain.ScheduleOverride, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ScheduleOverride
	for rows.Next() {
		o := &domain.ScheduleOverride{}
		var startTime, endTime *string
		if err := rows.Scan(&o.ID, &o.InstructorID, &o.Date, &o.IsDayOff, &startTime, &endTime, &o.CreatedAt); err != nil {
			return nil, err
		}
		if startTime != nil {
			o.StartTime = *startTime
		}
		if endTime != nil {
			o.EndTime = *endTime
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *ScheduleOverrideRepo) Delete(instructorID string, date time.Time) error {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `DELETE FROM instructor_schedule_overrides WHERE instructor_id = $1 AND date = $2`, instructorID, date)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
