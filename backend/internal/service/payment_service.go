package service

import (
	"context"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

// PaymentService turns a lesson's quoted price into an actual charge,
// splitting it between what the instructor is owed and the platform's fee.
// The instructor always receives exactly what they set as their rate — the
// fee is added on top of that, not deducted from it.
type PaymentService struct {
	payments    domain.PaymentRepository
	provider    domain.PaymentProvider
	feePercent  float64
	autoApprove bool // true while backed by the simulated internal/provider/manual
}

func NewPaymentService(payments domain.PaymentRepository, provider domain.PaymentProvider, feePercent float64, autoApprove bool) *PaymentService {
	return &PaymentService{payments: payments, provider: provider, feePercent: feePercent, autoApprove: autoApprove}
}

// ChargeForLesson creates the payment for a lesson that was just booked.
// With the simulated provider this resolves as approved immediately; a real
// provider would leave it pending until its webhook confirms the charge.
func (s *PaymentService) ChargeForLesson(ctx context.Context, lesson *domain.LessonRequest) (*domain.Payment, error) {
	instructorPrice := lesson.EstimatedPriceUYU
	fee := instructorPrice * s.feePercent / 100
	total := instructorPrice + fee

	checkout, err := s.provider.CreateCheckout(ctx, domain.CheckoutRequest{
		LessonRequestID: lesson.ID,
		AmountUYU:       total,
		Description:     "Clase de manejo",
	})
	if err != nil {
		return nil, err
	}

	payment := &domain.Payment{
		LessonRequestID:    lesson.ID,
		StudentID:          lesson.StudentID,
		InstructorID:       lesson.InstructorID,
		AmountUYU:          total,
		InstructorPriceUYU: instructorPrice,
		PlatformFeeUYU:     fee,
		ProviderRef:        checkout.ProviderRef,
		Status:             domain.PaymentPending,
	}
	if err := s.payments.Create(payment); err != nil {
		return nil, err
	}

	if s.autoApprove {
		if err := s.payments.UpdateStatus(payment.ID, domain.PaymentApproved, checkout.ProviderRef); err != nil {
			return nil, err
		}
		payment.Status = domain.PaymentApproved
	}

	return payment, nil
}

func (s *PaymentService) GetByLesson(lessonID string) (*domain.Payment, error) {
	return s.payments.GetByLessonID(lessonID)
}
