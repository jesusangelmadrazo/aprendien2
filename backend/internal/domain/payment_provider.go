package domain

import "context"

// CheckoutRequest is what a PaymentProvider needs to start a checkout for
// one completed lesson.
type CheckoutRequest struct {
	LessonRequestID string
	AmountUYU       float64
	Description     string
	SuccessURL      string
	FailureURL      string
	NotificationURL string
}

// CheckoutResult is what the frontend needs to send the student to pay.
type CheckoutResult struct {
	CheckoutURL string // where to redirect the student to complete payment
	ProviderRef string // provider-specific id (e.g. an MP preference id), persisted to reconcile later
}

// WebhookEvent is the provider-agnostic result of verifying an incoming
// payment notification.
type WebhookEvent struct {
	ProviderRef       string // matches CheckoutResult.ProviderRef, used to find our Payment row
	ProviderPaymentID string
	Status            PaymentStatus
}

// PaymentProvider abstracts a payment gateway (Mercado Pago, Stripe, etc.)
// behind a single contract so the rest of the app isn't coupled to any one
// of them. Implementations live under internal/provider/<name> and are
// selected at wiring time in cmd/api/main.go — swapping providers should
// never require touching PaymentService or the HTTP handlers.
type PaymentProvider interface {
	CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error)

	// ParseWebhook validates and normalizes an incoming provider webhook
	// request body into a WebhookEvent. Implementations must independently
	// re-verify the payment status against the provider's API rather than
	// trusting the webhook payload alone — that's a provider-specific
	// concern this interface deliberately hides.
	ParseWebhook(ctx context.Context, rawBody []byte) (*WebhookEvent, error)
}
