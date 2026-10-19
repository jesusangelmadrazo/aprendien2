package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

type AvailabilityHandler struct {
	availability *service.AvailabilityService
}

func NewAvailabilityHandler(availability *service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{availability: availability}
}

type createBlockRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"gte=0,lte=6"`
	StartTime string `json:"start_time" binding:"required"` // "HH:MM"
	EndTime   string `json:"end_time" binding:"required"`
}

func (h *AvailabilityHandler) CreateBlock(c *gin.Context) {
	instructorID := c.GetString("user_id")

	var req createBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	block, err := h.availability.CreateBlock(instructorID, time.Weekday(req.DayOfWeek), req.StartTime, req.EndTime)
	if errors.Is(err, service.ErrInvalidBlock) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create availability block"})
		return
	}
	c.JSON(http.StatusCreated, block)
}

func (h *AvailabilityHandler) ListMyBlocks(c *gin.Context) {
	instructorID := c.GetString("user_id")
	blocks, err := h.availability.ListBlocks(instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list availability"})
		return
	}
	c.JSON(http.StatusOK, blocks)
}

func (h *AvailabilityHandler) DeleteBlock(c *gin.Context) {
	instructorID := c.GetString("user_id")
	if err := h.availability.DeleteBlock(instructorID, c.Param("id")); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete block"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AvailabilityHandler) Slots(c *gin.Context) {
	instructorID := c.Param("id")
	daysAhead := 14
	if d, err := strconv.Atoi(c.Query("days")); err == nil && d > 0 && d <= 60 {
		daysAhead = d
	}

	slots, err := h.availability.AvailableSlots(instructorID, daysAhead)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute available slots"})
		return
	}
	c.JSON(http.StatusOK, slots)
}

// SlotsForDate powers the "filter by date, see everyone's openings" search:
// GET /api/instructors/slots?date=2026-09-20&instructor_id=optional
func (h *AvailabilityHandler) SlotsForDate(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query param is required (YYYY-MM-DD)"})
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be in YYYY-MM-DD format"})
		return
	}

	slots, err := h.availability.SlotsForDate(date, c.Query("instructor_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute available slots"})
		return
	}
	c.JSON(http.StatusOK, slots)
}

type setOverrideRequest struct {
	Date      string `json:"date" binding:"required"` // "YYYY-MM-DD"
	IsDayOff  bool   `json:"is_day_off"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (h *AvailabilityHandler) SetOverride(c *gin.Context) {
	instructorID := c.GetString("user_id")

	var req setOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be in YYYY-MM-DD format"})
		return
	}
	if !req.IsDayOff && (req.StartTime == "" || req.EndTime == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required unless is_day_off is true"})
		return
	}

	override, err := h.availability.SetOverride(instructorID, date, req.IsDayOff, req.StartTime, req.EndTime)
	if errors.Is(err, service.ErrInvalidBlock) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save override"})
		return
	}
	c.JSON(http.StatusOK, override)
}

func (h *AvailabilityHandler) ListMyOverrides(c *gin.Context) {
	instructorID := c.GetString("user_id")
	overrides, err := h.availability.ListUpcomingOverrides(instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list overrides"})
		return
	}
	c.JSON(http.StatusOK, overrides)
}

func (h *AvailabilityHandler) DeleteOverride(c *gin.Context) {
	instructorID := c.GetString("user_id")
	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be in YYYY-MM-DD format"})
		return
	}
	if err := h.availability.DeleteOverride(instructorID, date); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "override not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete override"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type bookSlotRequest struct {
	ScheduledAt     string  `json:"scheduled_at" binding:"required"` // RFC3339
	PickupLat       float64 `json:"pickup_lat" binding:"required"`
	PickupLon       float64 `json:"pickup_lon" binding:"required"`
	PickupAddress   string  `json:"pickup_address"`
	DurationMinutes int     `json:"duration_minutes"`
}

func (h *AvailabilityHandler) BookSlot(c *gin.Context) {
	studentID := c.GetString("user_id")
	instructorID := c.Param("id")

	var req bookSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be an RFC3339 timestamp"})
		return
	}
	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 60
	}

	result, err := h.availability.BookSlot(c.Request.Context(), studentID, instructorID, scheduledAt, req.PickupLat, req.PickupLon, req.PickupAddress, req.DurationMinutes)
	if errors.Is(err, service.ErrSlotNotAvailable) {
		c.JSON(http.StatusConflict, gin.H{"error": "ese horario ya no está disponible"})
		return
	}
	if errors.Is(err, service.ErrSlotInPast) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se puede reservar un horario pasado"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not book slot"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"lesson": result.Lesson, "payment": result.Payment})
}
