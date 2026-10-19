package domain

import "time"

type AccountType string

const (
	AccountSavings  AccountType = "caja_ahorro"
	AccountChecking AccountType = "cuenta_corriente"
)

// PayoutMethod is where an instructor wants their weekly earnings
// transferred. The actual transfer happens outside the app (a human does a
// real bank transfer); this is just the reference info for that.
type PayoutMethod struct {
	InstructorID          string
	BankName              string
	AccountType           AccountType
	AccountNumber         string
	AccountHolderName     string
	AccountHolderDocument string
	UpdatedAt             time.Time
}

type PayoutMethodRepository interface {
	Upsert(m *PayoutMethod) error
	GetByInstructor(instructorID string) (*PayoutMethod, error)
}

// Payout is a settled batch of an instructor's earnings — created by an
// admin once they've actually transferred the money, not by the app moving
// funds itself.
type Payout struct {
	ID           string
	InstructorID string
	AmountUYU    float64
	PeriodStart  time.Time
	PeriodEnd    time.Time
	CreatedBy    string
	CreatedAt    time.Time
}

type PayoutRepository interface {
	Create(p *Payout) error
	ListByInstructor(instructorID string) ([]*Payout, error)
}

// InstructorBalance is one row of the admin's "who's owed money" queue.
type InstructorBalance struct {
	InstructorID   string  `json:"instructor_id"`
	InstructorName string  `json:"instructor_name"`
	AmountUYU      float64 `json:"amount_uyu"`
}
