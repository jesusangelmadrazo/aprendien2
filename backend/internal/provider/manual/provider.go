// Package manual is a stand-in domain.PaymentProvider that simulates a
// successful checkout instead of calling a real gateway. It exists so the
// booking → payment flow can be built and tested end-to-end before Mercado
// Pago (or any other provider) sandbox credentials are available. Swapping
// it out later means writing a new package that satisfies
// domain.PaymentProvider and pointing cmd/api/main.go at it — nothing in
// PaymentService, the handlers, or the frontend needs to change.
package manual

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) CreateCheckout(_ context.Context, req domain.CheckoutRequest) (*domain.CheckoutResult, error) {
	ref := fmt.Sprintf("sim_%s_%d", req.LessonRequestID, time.Now().UnixNano())
	return &domain.CheckoutResult{
		CheckoutURL: req.SuccessURL,
		ProviderRef: ref,
	}, nil
}

// ParseWebhook is unused by the simulated flow (PaymentService approves the
// charge immediately instead of waiting for a callback) but is implemented
// so Provider satisfies domain.PaymentProvider.
func (p *Provider) ParseWebhook(_ context.Context, rawBody []byte) (*domain.WebhookEvent, error) {
	var evt domain.WebhookEvent
	if err := json.Unmarshal(rawBody, &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
