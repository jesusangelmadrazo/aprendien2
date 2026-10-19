package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
)

type UserHandler struct {
	users domain.UserRepository
}

func NewUserHandler(users domain.UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

// GetPublicInfo resolves a counterparty's name for someone who already knows
// their user_id from a shared record (a lesson, an offer) — e.g. an
// instructor showing the student's name on their reservations list. Not a
// general user directory: it only ever returns the display name, nothing else.
func (h *UserHandler) GetPublicInfo(c *gin.Context) {
	user, err := h.users.GetByID(c.Param("id"))
	if errors.Is(err, postgres.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "full_name": user.FullName})
}
