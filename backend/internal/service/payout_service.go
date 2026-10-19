package service

import (
	"errors"
	"time"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

var ErrNoPendingBalance = errors.New("instructor has no pending balance to pay out")

type WalletSummary struct {
	PendingUYU   float64          `json:"pending_uyu"`
	TotalPaidUYU float64          `json:"total_paid_uyu"`
	Payouts      []*domain.Payout `json:"payouts"`
}

// PayoutService is the instructor-earnings side of the marketplace: how much
// each instructor is owed, and settling that manually — an admin actually
// transfers the money (bank transfer, outside the app) and then records it
// here. There's no automated bank integration.
type PayoutService struct {
	payments      domain.PaymentRepository
	payouts       domain.PayoutRepository
	payoutMethods domain.PayoutMethodRepository
}

func NewPayoutService(payments domain.PaymentRepository, payouts domain.PayoutRepository, payoutMethods domain.PayoutMethodRepository) *PayoutService {
	return &PayoutService{payments: payments, payouts: payouts, payoutMethods: payoutMethods}
}

func (s *PayoutService) GetWallet(instructorID string) (*WalletSummary, error) {
	pending, err := s.payments.ListPendingByInstructor(instructorID)
	if err != nil {
		return nil, err
	}
	var pendingTotal float64
	for _, p := range pending {
		pendingTotal += p.InstructorPriceUYU
	}

	history, err := s.payouts.ListByInstructor(instructorID)
	if err != nil {
		return nil, err
	}
	var paidTotal float64
	for _, p := range history {
		paidTotal += p.AmountUYU
	}

	return &WalletSummary{PendingUYU: pendingTotal, TotalPaidUYU: paidTotal, Payouts: history}, nil
}

func (s *PayoutService) SetPayoutMethod(m *domain.PayoutMethod) error {
	return s.payoutMethods.Upsert(m)
}

func (s *PayoutService) GetPayoutMethod(instructorID string) (*domain.PayoutMethod, error) {
	return s.payoutMethods.GetByInstructor(instructorID)
}

func (s *PayoutService) ListPendingBalances() ([]domain.InstructorBalance, error) {
	return s.payments.ListPendingBalances()
}

// CreatePayout settles everything currently pending for one instructor into
// a single Payout record, then stamps those payments as paid. Call this
// after actually sending the money — it's a record of what happened, not
// what triggers the transfer.
func (s *PayoutService) CreatePayout(adminID, instructorID string) (*domain.Payout, error) {
	pending, err := s.payments.ListPendingByInstructor(instructorID)
	if err != nil {
		return nil, err
	}
	if len(pending) == 0 {
		return nil, ErrNoPendingBalance
	}

	var total float64
	periodStart := pending[0].CreatedAt
	periodEnd := pending[0].CreatedAt
	for _, p := range pending {
		total += p.InstructorPriceUYU
		if p.CreatedAt.Before(periodStart) {
			periodStart = p.CreatedAt
		}
		if p.CreatedAt.After(periodEnd) {
			periodEnd = p.CreatedAt
		}
	}
	if periodEnd.Before(time.Now()) {
		periodEnd = time.Now()
	}

	payout := &domain.Payout{
		InstructorID: instructorID,
		AmountUYU:    total,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		CreatedBy:    adminID,
	}
	if err := s.payouts.Create(payout); err != nil {
		return nil, err
	}
	if err := s.payments.MarkPaidForInstructor(instructorID, payout.ID); err != nil {
		return nil, err
	}
	return payout, nil
}
