package http

import (
	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/service"
	"github.com/jmadrazo/clase-drive-backend/internal/transport/ws"
)

type WSHandler struct {
	hub         *ws.Hub
	matcher     *service.Matcher
	instructors *service.InstructorService
}

func NewWSHandler(hub *ws.Hub, matcher *service.Matcher, instructors *service.InstructorService) *WSHandler {
	return &WSHandler{hub: hub, matcher: matcher, instructors: instructors}
}

func (h *WSHandler) Serve(c *gin.Context) {
	userID := c.GetString("user_id")
	ws.ServeWS(h.hub, h.matcher, h.instructors, c.Writer, c.Request, userID)
}
