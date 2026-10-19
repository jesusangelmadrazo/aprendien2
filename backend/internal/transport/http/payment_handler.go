package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
)

type PaymentHandler struct {
	payments domain.PaymentRepository
}

func NewPaymentHandler(payments domain.PaymentRepository) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

func (h *PaymentHandler) ListMine(c *gin.Context) {
	studentID := c.GetString("user_id")
	list, err := h.payments.ListByStudent(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list payments"})
		return
	}
	c.JSON(http.StatusOK, list)
}
