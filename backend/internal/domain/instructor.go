package domain

import "time"

type InstructorStatus string

const (
	InstructorOffline InstructorStatus = "offline"
	InstructorOnline  InstructorStatus = "online"
	InstructorBusy    InstructorStatus = "busy"
)

type InstructorProfile struct {
	UserID             string
	VehicleMake        string
	VehicleModel       string
	VehiclePlate       string
	VehicleYear        *int
	HourlyRateUYU      float64
	Bio                string
	Status             InstructorStatus
	VerificationStatus VerificationStatus
	AvgRating          float64
	TotalLessons       int
	LastKnownLat       *float64
	LastKnownLon       *float64
	LastSeenAt         *time.Time
	CreatedAt          time.Time
}

type NearbyInstructor struct {
	InstructorProfile
	FullName   string
	DistanceKm float64
	Lat        float64
	Lon        float64
}

// InstructorSummary is used to list instructors for scheduled booking —
// independent of whether they're online right now. Booking a slot for next
// Monday shouldn't require the instructor to be live-tracked at this exact
// moment the way real-time dispatch does.
type InstructorSummary struct {
	InstructorProfile
	FullName string
}

type InstructorRepository interface {
	Create(p *InstructorProfile) error
	GetByUserID(userID string) (*InstructorProfile, error)
	UpdateStatus(userID string, status InstructorStatus) error
	UpdateLastKnownLocation(userID string, lat, lon float64) error
	UpdateVerificationStatus(userID string, status VerificationStatus) error
	// ListAll only returns verified instructors — see DocumentService.
	ListAll() ([]*InstructorSummary, error)
}
