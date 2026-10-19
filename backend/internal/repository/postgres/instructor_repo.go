package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type InstructorRepo struct {
	pool *pgxpool.Pool
}

func NewInstructorRepo(pool *pgxpool.Pool) *InstructorRepo {
	return &InstructorRepo{pool: pool}
}

func (r *InstructorRepo) Create(p *domain.InstructorProfile) error {
	ctx := context.Background()
	return r.pool.QueryRow(ctx, `
		INSERT INTO instructor_profiles (user_id, vehicle_make, vehicle_model, vehicle_plate, vehicle_year, hourly_rate_uyu, bio)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`, p.UserID, p.VehicleMake, p.VehicleModel, p.VehiclePlate, p.VehicleYear, p.HourlyRateUYU, p.Bio).Scan(&p.CreatedAt)
}

func (r *InstructorRepo) GetByUserID(userID string) (*domain.InstructorProfile, error) {
	ctx := context.Background()
	p := &domain.InstructorProfile{}
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, vehicle_make, vehicle_model, vehicle_plate, vehicle_year, hourly_rate_uyu,
		       COALESCE(bio, ''), status, verification_status, avg_rating, total_lessons,
		       ST_Y(last_known_location::geometry), ST_X(last_known_location::geometry),
		       last_seen_at, created_at
		FROM instructor_profiles WHERE user_id = $1
	`, userID).Scan(
		&p.UserID, &p.VehicleMake, &p.VehicleModel, &p.VehiclePlate, &p.VehicleYear, &p.HourlyRateUYU,
		&p.Bio, &p.Status, &p.VerificationStatus, &p.AvgRating, &p.TotalLessons,
		&p.LastKnownLat, &p.LastKnownLon, &p.LastSeenAt, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *InstructorRepo) UpdateStatus(userID string, status domain.InstructorStatus) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `UPDATE instructor_profiles SET status = $1 WHERE user_id = $2`, status, userID)
	return err
}

func (r *InstructorRepo) UpdateVerificationStatus(userID string, status domain.VerificationStatus) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `UPDATE instructor_profiles SET verification_status = $1 WHERE user_id = $2`, status, userID)
	return err
}

func (r *InstructorRepo) UpdateLastKnownLocation(userID string, lat, lon float64) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		UPDATE instructor_profiles
		SET last_known_location = ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, last_seen_at = now()
		WHERE user_id = $3
	`, lon, lat, userID)
	return err
}

// ListAll returns every *verified* instructor profile regardless of online
// status, for discovery when booking a future scheduled slot. An instructor
// who hasn't passed document review yet shouldn't be bookable.
func (r *InstructorRepo) ListAll() ([]*domain.InstructorSummary, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT ip.user_id, ip.vehicle_make, ip.vehicle_model, ip.vehicle_plate, ip.vehicle_year, ip.hourly_rate_uyu,
		       COALESCE(ip.bio, ''), ip.status, ip.verification_status, ip.avg_rating, ip.total_lessons,
		       ST_Y(ip.last_known_location::geometry), ST_X(ip.last_known_location::geometry),
		       ip.last_seen_at, ip.created_at, u.full_name
		FROM instructor_profiles ip
		JOIN users u ON u.id = ip.user_id
		WHERE ip.verification_status = 'approved'
		ORDER BY u.full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.InstructorSummary
	for rows.Next() {
		s := &domain.InstructorSummary{}
		if err := rows.Scan(
			&s.UserID, &s.VehicleMake, &s.VehicleModel, &s.VehiclePlate, &s.VehicleYear, &s.HourlyRateUYU,
			&s.Bio, &s.Status, &s.VerificationStatus, &s.AvgRating, &s.TotalLessons,
			&s.LastKnownLat, &s.LastKnownLon, &s.LastSeenAt, &s.CreatedAt, &s.FullName,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
