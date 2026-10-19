package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

// ErrSlotTaken means the unique (instructor_id, scheduled_at) index rejected
// the insert — someone else booked that exact slot first.
var ErrSlotTaken = errors.New("slot already booked")

type LessonRepo struct {
	pool *pgxpool.Pool
}

func NewLessonRepo(pool *pgxpool.Pool) *LessonRepo {
	return &LessonRepo{pool: pool}
}

// Create inserts either an on-demand request (InstructorID nil, ScheduledAt
// nil — the Matcher assigns an instructor later) or a directly booked slot
// (both set upfront). The DB's unique partial index is what actually
// prevents two students from booking the same slot concurrently; this just
// translates that constraint violation into a typed error.
func (r *LessonRepo) Create(l *domain.LessonRequest) error {
	ctx := context.Background()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO lesson_requests (student_id, instructor_id, status, pickup_location, pickup_address_text, duration_minutes, estimated_price_uyu, scheduled_at)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, $6, $7, $8, $9)
		RETURNING id, requested_at
	`, l.StudentID, l.InstructorID, l.Status, l.PickupLon, l.PickupLat, l.PickupAddressText, l.DurationMinutes, l.EstimatedPriceUYU, l.ScheduledAt).
		Scan(&l.ID, &l.RequestedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSlotTaken
		}
		return err
	}
	return nil
}

func (r *LessonRepo) GetByID(id string) (*domain.LessonRequest, error) {
	ctx := context.Background()
	l := &domain.LessonRequest{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, student_id, instructor_id, status,
		       ST_Y(pickup_location::geometry), ST_X(pickup_location::geometry),
		       COALESCE(pickup_address_text, ''), duration_minutes, estimated_price_uyu, final_price_uyu,
		       scheduled_at, requested_at, accepted_at, started_at, completed_at, cancelled_at, COALESCE(cancel_reason, '')
		FROM lesson_requests WHERE id = $1
	`, id).Scan(
		&l.ID, &l.StudentID, &l.InstructorID, &l.Status,
		&l.PickupLat, &l.PickupLon,
		&l.PickupAddressText, &l.DurationMinutes, &l.EstimatedPriceUYU, &l.FinalPriceUYU,
		&l.ScheduledAt, &l.RequestedAt, &l.AcceptedAt, &l.StartedAt, &l.CompletedAt, &l.CancelledAt, &l.CancelReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

// AssignInstructor atomically assigns an instructor only if the request is still
// searching for one, preventing double-assignment under concurrent offers.
func (r *LessonRepo) AssignInstructor(lessonID, instructorID string) (bool, error) {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `
		UPDATE lesson_requests
		SET instructor_id = $1, status = 'accepted', accepted_at = now()
		WHERE id = $2 AND status = 'searching'
	`, instructorID, lessonID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *LessonRepo) UpdateStatus(lessonID string, status domain.LessonStatus) error {
	ctx := context.Background()
	var timestampCol string
	switch status {
	case domain.LessonInProgress:
		timestampCol = "started_at = now(),"
	case domain.LessonCompleted:
		timestampCol = "completed_at = now(),"
	case domain.LessonCancelled:
		timestampCol = "cancelled_at = now(),"
	}
	_, err := r.pool.Exec(ctx, `UPDATE lesson_requests SET `+timestampCol+` status = $1 WHERE id = $2`, status, lessonID)
	return err
}

func (r *LessonRepo) Cancel(lessonID, reason string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		UPDATE lesson_requests SET status = 'cancelled', cancelled_at = now(), cancel_reason = $1 WHERE id = $2
	`, reason, lessonID)
	return err
}

func (r *LessonRepo) MarkNoInstructorFound(lessonID string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `UPDATE lesson_requests SET status = 'no_instructor_found' WHERE id = $1`, lessonID)
	return err
}

func (r *LessonRepo) ListByUser(userID string, role domain.Role, status string) ([]*domain.LessonRequest, error) {
	ctx := context.Background()
	col := "student_id"
	if role == domain.RoleInstructor {
		col = "instructor_id"
	}
	query := `
		SELECT id, student_id, instructor_id, status,
		       ST_Y(pickup_location::geometry), ST_X(pickup_location::geometry),
		       COALESCE(pickup_address_text, ''), duration_minutes, estimated_price_uyu, final_price_uyu,
		       scheduled_at, requested_at, accepted_at, started_at, completed_at, cancelled_at, COALESCE(cancel_reason, '')
		FROM lesson_requests WHERE ` + col + ` = $1`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY requested_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.LessonRequest
	for rows.Next() {
		l := &domain.LessonRequest{}
		if err := rows.Scan(
			&l.ID, &l.StudentID, &l.InstructorID, &l.Status,
			&l.PickupLat, &l.PickupLon,
			&l.PickupAddressText, &l.DurationMinutes, &l.EstimatedPriceUYU, &l.FinalPriceUYU,
			&l.ScheduledAt, &l.RequestedAt, &l.AcceptedAt, &l.StartedAt, &l.CompletedAt, &l.CancelledAt, &l.CancelReason,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// ListScheduledByInstructor returns booked (non-cancelled) slots for an
// instructor within [from, to], used to exclude already-taken slots when
// computing available ones.
func (r *LessonRepo) ListScheduledByInstructor(instructorID string, from, to time.Time) ([]*domain.LessonRequest, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT id, scheduled_at
		FROM lesson_requests
		WHERE instructor_id = $1 AND scheduled_at IS NOT NULL AND scheduled_at BETWEEN $2 AND $3 AND status != 'cancelled'
	`, instructorID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.LessonRequest
	for rows.Next() {
		l := &domain.LessonRequest{}
		if err := rows.Scan(&l.ID, &l.ScheduledAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *LessonRepo) CompletionStats(studentID string) (int, int, error) {
	ctx := context.Background()
	var lessonsCompleted, minutesCompleted int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(duration_minutes), 0)
		FROM lesson_requests WHERE student_id = $1 AND status = 'completed'
	`, studentID).Scan(&lessonsCompleted, &minutesCompleted)
	return lessonsCompleted, minutesCompleted, err
}

func (r *LessonRepo) CreateOffer(o *domain.LessonOffer) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO lesson_offers (lesson_request_id, instructor_id, sequence_order, result)
		VALUES ($1, $2, $3, $4)
		RETURNING id, offered_at
	`, o.LessonRequestID, o.InstructorID, o.SequenceOrder, o.Result).Scan(&o.ID, &o.OfferedAt)
}

func (r *LessonRepo) GetOffer(id string) (*domain.LessonOffer, error) {
	ctx := context.Background()
	o := &domain.LessonOffer{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, lesson_request_id, instructor_id, sequence_order, result, offered_at, responded_at
		FROM lesson_offers WHERE id = $1
	`, id).Scan(&o.ID, &o.LessonRequestID, &o.InstructorID, &o.SequenceOrder, &o.Result, &o.OfferedAt, &o.RespondedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (r *LessonRepo) SetOfferResult(id string, result domain.OfferResult) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `UPDATE lesson_offers SET result = $1, responded_at = now() WHERE id = $2`, result, id)
	return err
}
