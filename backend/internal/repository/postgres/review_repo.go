package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) Create(rev *domain.Review) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO reviews (lesson_request_id, reviewer_id, reviewee_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, rev.LessonRequestID, rev.ReviewerID, rev.RevieweeID, rev.Rating, rev.Comment).Scan(&rev.ID, &rev.CreatedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE instructor_profiles SET
			avg_rating = (SELECT AVG(rating) FROM reviews WHERE reviewee_id = $1),
			total_lessons = total_lessons + 1
		WHERE user_id = $1
	`, rev.RevieweeID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ReviewRepo) ListByInstructor(instructorID string) ([]*domain.Review, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_request_id, reviewer_id, reviewee_id, rating, COALESCE(comment, ''), created_at
		FROM reviews WHERE reviewee_id = $1 ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Review
	for rows.Next() {
		rev := &domain.Review{}
		if err := rows.Scan(&rev.ID, &rev.LessonRequestID, &rev.ReviewerID, &rev.RevieweeID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}
