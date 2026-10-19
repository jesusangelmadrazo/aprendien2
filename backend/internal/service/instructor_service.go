package service

import (
	"context"
	"errors"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/redis"
)

var ErrNotVerified = errors.New("instructor has not been verified yet")

type InstructorService struct {
	instructors domain.InstructorRepository
	users       domain.UserRepository
	geo         *redis.GeoRepo
}

func NewInstructorService(instructors domain.InstructorRepository, users domain.UserRepository, geo *redis.GeoRepo) *InstructorService {
	return &InstructorService{instructors: instructors, users: users, geo: geo}
}

func (s *InstructorService) CreateProfile(p *domain.InstructorProfile) error {
	return s.instructors.Create(p)
}

func (s *InstructorService) GetProfile(userID string) (*domain.InstructorProfile, error) {
	return s.instructors.GetByUserID(userID)
}

// SetStatus toggles instructor availability. Going online publishes the instructor
// to Redis GEO for live matching; going offline/busy removes it from search.
// Going online requires the instructor to have passed document review first
// (license, vehicle registration, insurance — see DocumentService).
func (s *InstructorService) SetStatus(ctx context.Context, userID string, status domain.InstructorStatus, lat, lon *float64) error {
	if status == domain.InstructorOnline {
		profile, err := s.instructors.GetByUserID(userID)
		if err != nil {
			return err
		}
		if profile.VerificationStatus != domain.VerificationApproved {
			return ErrNotVerified
		}
	}
	if err := s.instructors.UpdateStatus(userID, status); err != nil {
		return err
	}
	if status == domain.InstructorOnline && lat != nil && lon != nil {
		return s.geo.SetLocation(ctx, userID, *lat, *lon)
	}
	return s.geo.RemoveInstructor(ctx, userID)
}

func (s *InstructorService) UpdateLocation(ctx context.Context, userID string, lat, lon float64) error {
	if err := s.instructors.UpdateLastKnownLocation(userID, lat, lon); err != nil {
		return err
	}
	return s.geo.SetLocation(ctx, userID, lat, lon)
}

// Nearby lists online instructors within radiusKm, closest first, for the
// student-facing map. Profiles that fail to load (e.g. a stale geo entry
// left behind by a crash) are skipped rather than failing the whole search.
func (s *InstructorService) Nearby(ctx context.Context, lat, lon, radiusKm float64) ([]domain.NearbyInstructor, error) {
	results, err := s.geo.SearchNearby(ctx, lat, lon, radiusKm, 20)
	if err != nil {
		return nil, err
	}

	out := make([]domain.NearbyInstructor, 0, len(results))
	for _, r := range results {
		profile, err := s.instructors.GetByUserID(r.InstructorID)
		if err != nil {
			continue
		}
		user, err := s.users.GetByID(r.InstructorID)
		if err != nil {
			continue
		}
		out = append(out, domain.NearbyInstructor{
			InstructorProfile: *profile,
			FullName:          user.FullName,
			DistanceKm:        r.DistanceKm,
			Lat:               r.Lat,
			Lon:               r.Lon,
		})
	}
	return out, nil
}

// ListAll returns every instructor regardless of online status, for
// browsing/booking a future slot — see domain.InstructorSummary.
func (s *InstructorService) ListAll() ([]*domain.InstructorSummary, error) {
	return s.instructors.ListAll()
}
