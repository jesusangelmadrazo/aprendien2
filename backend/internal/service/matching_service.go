package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	redisrepo "github.com/jmadrazo/clase-drive-backend/internal/repository/redis"
	"github.com/jmadrazo/clase-drive-backend/internal/transport/ws"
)

var ErrOfferNotPending = errors.New("offer is not pending a response")

const (
	firstSearchRadiusKm  = 5
	secondSearchRadiusKm = 10
	maxCandidates        = 5
	offerTimeout         = 15 * time.Second
)

// Matcher orchestrates one lesson request at a time via a goroutine, offering
// it to the nearest online instructor and falling through to the next
// candidate on rejection or timeout. It intentionally lives in memory (not
// Redis Pub/Sub): the backend runs as a single process for the MVP, so plain
// channels are simpler and faster than round-tripping through Redis. If this
// ever needs multiple backend instances, offer orchestration would need to
// move to a shared broker instead.
type Matcher struct {
	lessons     domain.LessonRepository
	instructors domain.InstructorRepository
	users       domain.UserRepository
	geo         *redisrepo.GeoRepo
	hub         *ws.Hub

	mu      sync.Mutex
	pending map[string]chan bool // offer ID -> response channel
}

func NewMatcher(lessons domain.LessonRepository, instructors domain.InstructorRepository, users domain.UserRepository, geo *redisrepo.GeoRepo, hub *ws.Hub) *Matcher {
	return &Matcher{
		lessons:     lessons,
		instructors: instructors,
		users:       users,
		geo:         geo,
		hub:         hub,
		pending:     make(map[string]chan bool),
	}
}

// RequestLesson kicks off matching for a freshly created lesson request.
// It returns immediately; progress is pushed to both parties over WebSocket.
func (m *Matcher) RequestLesson(lessonID string) {
	go m.run(lessonID)
}

func (m *Matcher) run(lessonID string) {
	ctx := context.Background()

	lesson, err := m.lessons.GetByID(lessonID)
	if err != nil {
		slog.Error("matcher: could not load lesson", "lesson_id", lessonID, "error", err)
		return
	}

	candidates, err := m.geo.SearchNearby(ctx, lesson.PickupLat, lesson.PickupLon, firstSearchRadiusKm, maxCandidates)
	if err == nil && len(candidates) == 0 {
		candidates, err = m.geo.SearchNearby(ctx, lesson.PickupLat, lesson.PickupLon, secondSearchRadiusKm, maxCandidates)
	}
	if err != nil {
		slog.Error("matcher: geo search failed", "lesson_id", lessonID, "error", err)
	}

	for i, candidate := range candidates {
		if m.tryOffer(ctx, lesson, candidate, i+1) {
			return
		}
	}

	if err := m.lessons.MarkNoInstructorFound(lessonID); err != nil {
		slog.Error("matcher: could not mark no_instructor_found", "lesson_id", lessonID, "error", err)
	}
	m.hub.SendToUser(ctx, lesson.StudentID, ws.Event{
		Type:    "no_instructor_found",
		Payload: map[string]string{"lesson_request_id": lesson.ID},
	})
}

// tryOffer offers the lesson to one candidate and returns true once the
// lesson has been successfully assigned (ending the search).
func (m *Matcher) tryOffer(ctx context.Context, lesson *domain.LessonRequest, candidate redisrepo.NearbyResult, sequence int) bool {
	offer := &domain.LessonOffer{
		LessonRequestID: lesson.ID,
		InstructorID:    candidate.InstructorID,
		SequenceOrder:   sequence,
		Result:          domain.OfferPending,
	}
	if err := m.lessons.CreateOffer(offer); err != nil {
		slog.Error("matcher: could not create offer", "lesson_id", lesson.ID, "error", err)
		return false
	}

	respCh := make(chan bool, 1)
	m.mu.Lock()
	m.pending[offer.ID] = respCh
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.pending, offer.ID)
		m.mu.Unlock()
	}()

	// Take the candidate out of search while they're deciding so they can't
	// receive a second offer for a different request in the meantime.
	_ = m.instructors.UpdateStatus(candidate.InstructorID, domain.InstructorBusy)
	_ = m.geo.RemoveInstructor(ctx, candidate.InstructorID)

	m.hub.SendToUser(ctx, candidate.InstructorID, ws.Event{
		Type: "lesson_offer",
		Payload: map[string]interface{}{
			"lesson_request_id": lesson.ID,
			"offer_id":          offer.ID,
			"pickup_address":    lesson.PickupAddressText,
			"pickup_lat":        lesson.PickupLat,
			"pickup_lon":        lesson.PickupLon,
			"duration_minutes":  lesson.DurationMinutes,
			"distance_km":       candidate.DistanceKm,
			"expires_in_s":      int(offerTimeout.Seconds()),
		},
	})

	accepted := false
	select {
	case accepted = <-respCh:
	case <-time.After(offerTimeout):
		_ = m.lessons.SetOfferResult(offer.ID, domain.OfferTimedOut)
		m.hub.SendToUser(ctx, candidate.InstructorID, ws.Event{
			Type:    "offer_expired",
			Payload: map[string]string{"offer_id": offer.ID},
		})
	}

	if !accepted {
		_ = m.instructors.UpdateStatus(candidate.InstructorID, domain.InstructorOnline)
		_ = m.geo.SetLocation(ctx, candidate.InstructorID, candidate.Lat, candidate.Lon)
		return false
	}

	assigned, err := m.lessons.AssignInstructor(lesson.ID, candidate.InstructorID)
	if err != nil {
		slog.Error("matcher: could not assign instructor", "lesson_id", lesson.ID, "error", err)
		return false
	}
	if !assigned {
		// Lost a race (shouldn't happen since only one Matcher run exists per
		// lesson, but guards against manual DB edits during development).
		_ = m.instructors.UpdateStatus(candidate.InstructorID, domain.InstructorOnline)
		_ = m.geo.SetLocation(ctx, candidate.InstructorID, candidate.Lat, candidate.Lon)
		return false
	}

	_ = m.lessons.SetOfferResult(offer.ID, domain.OfferAccepted)

	instructor, err := m.users.GetByID(candidate.InstructorID)
	instructorName := ""
	if err == nil {
		instructorName = instructor.FullName
	}
	profile, _ := m.instructors.GetByUserID(candidate.InstructorID)

	m.hub.SendToUser(ctx, lesson.StudentID, ws.Event{
		Type: "instructor_assigned",
		Payload: map[string]interface{}{
			"lesson_request_id": lesson.ID,
			"instructor_name":   instructorName,
			"vehicle":           vehicleSummary(profile),
			"distance_km":       candidate.DistanceKm,
		},
	})
	return true
}

func vehicleSummary(p *domain.InstructorProfile) string {
	if p == nil {
		return ""
	}
	return p.VehicleMake + " " + p.VehicleModel + " - " + p.VehiclePlate
}

// RespondToOffer is called from the WebSocket read pump (or a REST fallback)
// when an instructor accepts or rejects a pending offer. A response to an
// offer that already timed out or was answered is simply ignored.
func (m *Matcher) RespondToOffer(offerID string, accepted bool) error {
	m.mu.Lock()
	ch, ok := m.pending[offerID]
	m.mu.Unlock()
	if !ok {
		return ErrOfferNotPending
	}
	select {
	case ch <- accepted:
	default:
	}
	if !accepted {
		_ = m.lessons.SetOfferResult(offerID, domain.OfferRejected)
	}
	return nil
}
