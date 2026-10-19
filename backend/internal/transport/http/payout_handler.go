package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

type PayoutHandler struct {
	payouts *service.PayoutService
}

func NewPayoutHandler(payouts *service.PayoutService) *PayoutHandler {
	return &PayoutHandler{payouts: payouts}
}

func (h *PayoutHandler) GetMyWallet(c *gin.Context) {
	instructorID := c.GetString("user_id")
	wallet, err := h.payouts.GetWallet(instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load wallet"})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

type setPayoutMethodRequest struct {
	BankName              string `json:"bank_name" binding:"required"`
	AccountType           string `json:"account_type" binding:"required,oneof=caja_ahorro cuenta_corriente"`
	AccountNumber         string `json:"account_number" binding:"required"`
	AccountHolderName     string `json:"account_holder_name" binding:"required"`
	AccountHolderDocument string `json:"account_holder_document" binding:"required"`
}

func (h *PayoutHandler) SetMyPayoutMethod(c *gin.Context) {
	instructorID := c.GetString("user_id")

	var req setPayoutMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method := &domain.PayoutMethod{
		InstructorID:          instructorID,
		BankName:              req.BankName,
		AccountType:           domain.AccountType(req.AccountType),
		AccountNumber:         req.AccountNumber,
		AccountHolderName:     req.AccountHolderName,
		AccountHolderDocument: req.AccountHolderDocument,
	}
	if err := h.payouts.SetPayoutMethod(method); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save payout method"})
		return
	}
	c.JSON(http.StatusOK, method)
}

func (h *PayoutHandler) GetMyPayoutMethod(c *gin.Context) {
	instructorID := c.GetString("user_id")
	method, err := h.payouts.GetPayoutMethod(instructorID)
	if errors.Is(err, postgres.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "no payout method set"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch payout method"})
		return
	}
	c.JSON(http.StatusOK, method)
}

func (h *PayoutHandler) ListPendingBalances(c *gin.Context) {
	balances, err := h.payouts.ListPendingBalances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list pending balances"})
		return
	}
	c.JSON(http.StatusOK, balances)
}

func (h *PayoutHandler) CreatePayout(c *gin.Context) {
	adminID := c.GetString("user_id")
	instructorID := c.Param("instructorId")

	payout, err := h.payouts.CreatePayout(adminID, instructorID)
	if errors.Is(err, service.ErrNoPendingBalance) {
		c.JSON(http.StatusConflict, gin.H{"error": "ese instructor no tiene saldo pendiente"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create payout"})
		return
	}
	c.JSON(http.StatusCreated, payout)
}
