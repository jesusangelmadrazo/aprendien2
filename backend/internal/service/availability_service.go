package service

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"time"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	"github.com/jmadrazo/clase-drive-backend/internal/transport/ws"
)

var (
	ErrInvalidBlock     = errors.New("start_time must be before end_time")
	ErrSlotNotAvailable = errors.New("that slot is not available")
	ErrSlotInPast       = errors.New("cannot book a slot in the past")
)

// defaultSlotMinutes is the fixed lesson length used to slice an
// instructor's recurring availability blocks into concrete bookable slots.
const defaultSlotMinutes = 60

var uruguayLocation = func() *time.Location {
	loc, err := time.LoadLocation("America/Montevideo")
	if err != nil {
		slog.Warn("could not load America/Montevideo timezone, falling back to UTC", "error", err)
		return time.UTC
	}
	return loc
}()

type AvailabilitySlot struct {
	InstructorID string    `json:"instructor_id"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
}

type AvailabilityService struct {
	availability domain.AvailabilityRepository
	overrides    domain.ScheduleOverrideRepository
	lessons      domain.LessonRepository
	instructors  domain.InstructorRepository
	payments     *PaymentService
	hub          *ws.Hub
}

func NewAvailabilityService(availability domain.AvailabilityRepository, overrides domain.ScheduleOverrideRepository, lessons domain.LessonRepository, instructors domain.InstructorRepository, payments *PaymentService, hub *ws.Hub) *AvailabilityService {
	return &AvailabilityService{availability: availability, overrides: overrides, lessons: lessons, instructors: instructors, payments: payments, hub: hub}
}

// BookingResult bundles the confirmed lesson with what the student was
// actually charged for it.
type BookingResult struct {
	Lesson  *domain.LessonRequest
	Payment *domain.Payment
}

func (s *AvailabilityService) CreateBlock(instructorID string, dayOfWeek time.Weekday, startTime, endTime string) (*domain.AvailabilityBlock, error) {
	if startTime >= endTime {
		return nil, ErrInvalidBlock
	}
	block := &domain.AvailabilityBlock{InstructorID: instructorID, DayOfWeek: dayOfWeek, StartTime: startTime, EndTime: endTime}
	if err := s.availability.Create(block); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *AvailabilityService) ListBlocks(instructorID string) ([]*domain.AvailabilityBlock, error) {
	return s.availability.ListByInstructor(instructorID)
}

func (s *AvailabilityService) DeleteBlock(instructorID, blockID string) error {
	return s.availability.Delete(blockID, instructorID)
}

// SetOverride pins one calendar date to a day off or a custom time range,
// replacing the recurring weekly pattern for that date only.
func (s *AvailabilityService) SetOverride(instructorID string, date time.Time, isDayOff bool, startTime, endTime string) (*domain.ScheduleOverride, error) {
	if !isDayOff && startTime >= endTime {
		return nil, ErrInvalidBlock
	}
	override := &domain.ScheduleOverride{
		InstructorID: instructorID,
		Date:         truncateToDay(date),
		IsDayOff:     isDayOff,
		StartTime:    startTime,
		EndTime:      endTime,
	}
	if err := s.overrides.Upsert(override); err != nil {
		return nil, err
	}
	return override, nil
}

func (s *AvailabilityService) ListUpcomingOverrides(instructorID string) ([]*domain.ScheduleOverride, error) {
	return s.overrides.ListUpcomingByInstructor(instructorID, truncateToDay(time.Now().In(uruguayLocation)))
}

func (s *AvailabilityService) DeleteOverride(instructorID string, date time.Time) error {
	return s.overrides.Delete(instructorID, truncateToDay(date))
}

// AvailableSlots expands the instructor's recurring weekly blocks into
// concrete slots over the next daysAhead days, excluding ones already
// booked and ones already in the past.
func (s *AvailabilityService) AvailableSlots(instructorID string, daysAhead int) ([]AvailabilitySlot, error) {
	now := time.Now().In(uruguayLocation)
	return s.SlotsInRange(instructorID, now, now.AddDate(0, 0, daysAhead))
}

// SlotsInRange is the same expansion as AvailableSlots but bounded by an
// explicit [from, to) window instead of "the next N days from now" — used
// to pull just one specific day's slots when aggregating across instructors.
// A date with a schedule override uses only that override (day off = no
// slots at all that day, custom range = only that range) instead of the
// recurring weekly blocks.
func (s *AvailabilityService) SlotsInRange(instructorID string, from, to time.Time) ([]AvailabilitySlot, error) {
	blocks, err := s.availability.ListByInstructor(instructorID)
	if err != nil {
		return nil, err
	}

	from = from.In(uruguayLocation)
	to = to.In(uruguayLocation)
	now := time.Now().In(uruguayLocation)

	overrides, err := s.overrides.ListByInstructorInRange(instructorID, truncateToDay(from), truncateToDay(to))
	if err != nil {
		return nil, err
	}
	overrideByDate := make(map[string]*domain.ScheduleOverride, len(overrides))
	for _, o := range overrides {
		overrideByDate[o.Date.Format("2006-01-02")] = o
	}

	booked, err := s.lessons.ListScheduledByInstructor(instructorID, from, to)
	if err != nil {
		return nil, err
	}
	bookedAt := make(map[int64]bool, len(booked))
	for _, l := range booked {
		if l.ScheduledAt != nil {
			bookedAt[l.ScheduledAt.Unix()] = true
		}
	}

	slotDuration := time.Duration(defaultSlotMinutes) * time.Minute
	var slots []AvailabilitySlot
	for date := truncateToDay(from); !date.After(to); date = date.AddDate(0, 0, 1) {
		ranges := effectiveRangesForDate(date, blocks, overrideByDate[date.Format("2006-01-02")])
		for _, r := range ranges {
			start, err := combineDateAndTime(date, r.start)
			if err != nil {
				continue
			}
			end, err := combineDateAndTime(date, r.end)
			if err != nil {
				continue
			}
			for t := start; !t.Add(slotDuration).After(end); t = t.Add(slotDuration) {
				if t.Before(now) || t.Before(from) || t.After(to) || bookedAt[t.Unix()] {
					continue
				}
				slots = append(slots, AvailabilitySlot{InstructorID: instructorID, Start: t, End: t.Add(slotDuration)})
			}
		}
	}
	return slots, nil
}

// SlotWithInstructor is one bookable slot annotated with enough instructor
// info to render it in a cross-instructor search without a second lookup.
type SlotWithInstructor struct {
	AvailabilitySlot
	InstructorName string  `json:"instructor_name"`
	VehicleMake    string  `json:"vehicle_make"`
	VehicleModel   string  `json:"vehicle_model"`
	HourlyRateUYU  float64 `json:"hourly_rate_uyu"`
}

// SlotsForDate lists every available slot on one calendar date, across all
// instructors (or just one, if instructorID is given) — the "filter by date,
// see everyone's openings" view.
func (s *AvailabilityService) SlotsForDate(date time.Time, instructorID string) ([]SlotWithInstructor, error) {
	all, err := s.instructors.ListAll()
	if err != nil {
		return nil, err
	}

	targets := all
	if instructorID != "" {
		targets = nil
		for _, inst := range all {
			if inst.UserID == instructorID {
				targets = append(targets, inst)
				break
			}
		}
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, uruguayLocation)
	dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)

	var out []SlotWithInstructor
	for _, inst := range targets {
		slots, err := s.SlotsInRange(inst.UserID, dayStart, dayEnd)
		if err != nil {
			continue
		}
		for _, slot := range slots {
			out = append(out, SlotWithInstructor{
				AvailabilitySlot: slot,
				InstructorName:   inst.FullName,
				VehicleMake:      inst.VehicleMake,
				VehicleModel:     inst.VehicleModel,
				HourlyRateUYU:    inst.HourlyRateUYU,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BookSlot directly confirms a lesson with a specific instructor at a
// specific time (first request to land wins — the DB's unique index is the
// actual source of truth for that, this just validates intent beforehand).
func (s *AvailabilityService) BookSlot(ctx context.Context, studentID, instructorID string, scheduledAt time.Time, pickupLat, pickupLon float64, pickupAddress string, durationMinutes int) (*BookingResult, error) {
	if scheduledAt.Before(time.Now()) {
		return nil, ErrSlotInPast
	}

	fits, err := s.fitsSchedule(instructorID, scheduledAt)
	if err != nil {
		return nil, err
	}
	if !fits {
		return nil, ErrSlotNotAvailable
	}

	profile, err := s.instructors.GetByUserID(instructorID)
	if err != nil {
		return nil, err
	}
	price := profile.HourlyRateUYU * float64(durationMinutes) / 60.0

	lesson := &domain.LessonRequest{
		StudentID:         studentID,
		InstructorID:      &instructorID,
		Status:            domain.LessonAccepted,
		PickupLat:         pickupLat,
		PickupLon:         pickupLon,
		PickupAddressText: pickupAddress,
		DurationMinutes:   durationMinutes,
		EstimatedPriceUYU: price,
		ScheduledAt:       &scheduledAt,
	}
	if err := s.lessons.Create(lesson); err != nil {
		if errors.Is(err, postgres.ErrSlotTaken) {
			return nil, ErrSlotNotAvailable
		}
		return nil, err
	}

	// Charge after the slot is secured, not before: booking is what actually
	// protects the slot (via the DB's unique index), so it should win the
	// race even if the simulated/real charge takes a moment.
	payment, err := s.payments.ChargeForLesson(ctx, lesson)
	if err != nil {
		return nil, err
	}

	s.hub.SendToUser(ctx, instructorID, ws.Event{
		Type: "lesson_scheduled",
		Payload: map[string]interface{}{
			"lesson_request_id": lesson.ID,
			"scheduled_at":      scheduledAt,
			"pickup_address":    pickupAddress,
		},
	})
	return &BookingResult{Lesson: lesson, Payment: payment}, nil
}

// fitsSchedule checks a specific instant against that date's *effective*
// schedule — an override (day off or custom range) if one exists for that
// date, otherwise the recurring weekly blocks.
func (s *AvailabilityService) fitsSchedule(instructorID string, t time.Time) (bool, error) {
	local := t.In(uruguayLocation)

	override, err := s.overrides.GetByInstructorAndDate(instructorID, truncateToDay(local))
	if err != nil && !errors.Is(err, postgres.ErrNotFound) {
		return false, err
	}
	if override != nil {
		if override.IsDayOff {
			return false, nil
		}
		return minutesOfDayWithin(local, override.StartTime, override.EndTime), nil
	}

	blocks, err := s.availability.ListByInstructor(instructorID)
	if err != nil {
		return false, err
	}
	for _, b := range blocks {
		if local.Weekday() == b.DayOfWeek && minutesOfDayWithin(local, b.StartTime, b.EndTime) {
			return true, nil
		}
	}
	return false, nil
}

type timeRange struct{ start, end string }

// effectiveRangesForDate is the override-aware version of "which blocks
// apply on this date": an override (day off or custom range) replaces the
// recurring weekly blocks entirely for that date.
func effectiveRangesForDate(date time.Time, blocks []*domain.AvailabilityBlock, override *domain.ScheduleOverride) []timeRange {
	if override != nil {
		if override.IsDayOff {
			return nil
		}
		return []timeRange{{start: override.StartTime, end: override.EndTime}}
	}
	var out []timeRange
	for _, b := range blocks {
		if date.Weekday() == b.DayOfWeek {
			out = append(out, timeRange{start: b.StartTime, end: b.EndTime})
		}
	}
	return out
}

func minutesOfDayWithin(t time.Time, startHHMM, endHHMM string) bool {
	startMinutes, err := minutesFromHHMM(startHHMM)
	if err != nil {
		return false
	}
	endMinutes, err := minutesFromHHMM(endHHMM)
	if err != nil {
		return false
	}
	minutesOfDay := t.Hour()*60 + t.Minute()
	return minutesOfDay >= startMinutes && minutesOfDay < endMinutes
}

func combineDateAndTime(date time.Time, hhmm string) (time.Time, error) {
	t, err := time.ParseInLocation("15:04", hhmm, uruguayLocation)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, uruguayLocation), nil
}

func minutesFromHHMM(hhmm string) (int, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}
