package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

// defaultHourlyRateUYU prices a lesson request before an instructor is
// assigned. It's a flat MVP placeholder; a per-instructor rate only exists
// once matching picks a candidate.
const defaultHourlyRateUYU = 800.0

type LessonHandler struct {
	lessons domain.LessonRepository
	matcher *service.Matcher
}

func NewLessonHandler(lessons domain.LessonRepository, matcher *service.Matcher) *LessonHandler {
	return &LessonHandler{lessons: lessons, matcher: matcher}
}

type createLessonRequest struct {
	PickupLat       float64 `json:"pickup_lat" binding:"required"`
	PickupLon       float64 `json:"pickup_lon" binding:"required"`
	PickupAddress   string  `json:"pickup_address"`
	DurationMinutes int     `json:"duration_minutes"`
}

func (h *LessonHandler) Create(c *gin.Context) {
	studentID := c.GetString("user_id")

	var req createLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 60
	}

	lesson := &domain.LessonRequest{
		StudentID:         studentID,
		Status:            domain.LessonSearching,
		PickupLat:         req.PickupLat,
		PickupLon:         req.PickupLon,
		PickupAddressText: req.PickupAddress,
		DurationMinutes:   req.DurationMinutes,
		EstimatedPriceUYU: defaultHourlyRateUYU * float64(req.DurationMinutes) / 60.0,
	}
	if err := h.lessons.Create(lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create lesson request"})
		return
	}

	h.matcher.RequestLesson(lesson.ID)
	c.JSON(http.StatusCreated, lesson)
}

func (h *LessonHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	lesson, err := h.lessons.GetByID(c.Param("id"))
	if errors.Is(err, postgres.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch lesson"})
		return
	}
	if lesson.StudentID != userID && (lesson.InstructorID == nil || *lesson.InstructorID != userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, lesson)
}

func (h *LessonHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	role := domain.Role(c.GetString("role"))
	status := c.Query("status")

	list, err := h.lessons.ListByUser(userID, role, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list lessons"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type respondOfferRequest struct {
	Accept bool `json:"accept"`
}

// RespondOffer is a REST fallback for instructors without an open WebSocket
// connection; the same action is also available as a "lesson_offer_response"
// WS message.
func (h *LessonHandler) RespondOffer(c *gin.Context) {
	var req respondOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.matcher.RespondToOffer(c.Param("offer_id"), req.Accept); err != nil {
		c.JSON(http.StatusGone, gin.H{"error": "offer is no longer pending"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Start marks the lesson as in progress once the instructor has arrived,
// per the MVP flow (no separate "en route" endpoint yet).
func (h *LessonHandler) Start(c *gin.Context) {
	h.transition(c, domain.LessonInProgress)
}

func (h *LessonHandler) Complete(c *gin.Context) {
	h.transition(c, domain.LessonCompleted)
}

func (h *LessonHandler) transition(c *gin.Context, status domain.LessonStatus) {
	if err := h.lessons.UpdateStatus(c.Param("id"), status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update lesson status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

type cancelLessonRequest struct {
	Reason string `json:"reason"`
}

func (h *LessonHandler) Cancel(c *gin.Context) {
	var req cancelLessonRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.lessons.Cancel(c.Param("id"), req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not cancel lesson"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": domain.LessonCancelled})
}
