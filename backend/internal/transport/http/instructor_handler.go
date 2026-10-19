package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

type InstructorHandler struct {
	instructors *service.InstructorService
}

func NewInstructorHandler(instructors *service.InstructorService) *InstructorHandler {
	return &InstructorHandler{instructors: instructors}
}

type createProfileRequest struct {
	VehicleMake   string  `json:"vehicle_make" binding:"required"`
	VehicleModel  string  `json:"vehicle_model" binding:"required"`
	VehiclePlate  string  `json:"vehicle_plate" binding:"required"`
	VehicleYear   *int    `json:"vehicle_year"`
	HourlyRateUYU float64 `json:"hourly_rate_uyu" binding:"required,gt=0"`
	Bio           string  `json:"bio"`
}

func (h *InstructorHandler) CreateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req createProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := &domain.InstructorProfile{
		UserID:        userID,
		VehicleMake:   req.VehicleMake,
		VehicleModel:  req.VehicleModel,
		VehiclePlate:  req.VehiclePlate,
		VehicleYear:   req.VehicleYear,
		HourlyRateUYU: req.HourlyRateUYU,
		Bio:           req.Bio,
	}
	if err := h.instructors.CreateProfile(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create profile"})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *InstructorHandler) GetMyProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	p, err := h.instructors.GetProfile(userID)
	if errors.Is(err, postgres.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch profile"})
		return
	}

	c.JSON(http.StatusOK, p)
}

type updateStatusRequest struct {
	Status string   `json:"status" binding:"required,oneof=online offline busy"`
	Lat    *float64 `json:"lat"`
	Lon    *float64 `json:"lon"`
}

func (h *InstructorHandler) UpdateStatus(c *gin.Context) {
	userID := c.GetString("user_id")

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.instructors.SetStatus(c.Request.Context(), userID, domain.InstructorStatus(req.Status), req.Lat, req.Lon); err != nil {
		if errors.Is(err, service.ErrNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"error": "tu cuenta todavía no fue verificada por un admin"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

type updateLocationRequest struct {
	Lat float64 `json:"lat" binding:"required"`
	Lon float64 `json:"lon" binding:"required"`
}

func (h *InstructorHandler) UpdateLocation(c *gin.Context) {
	userID := c.GetString("user_id")

	var req updateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.instructors.UpdateLocation(c.Request.Context(), userID, req.Lat, req.Lon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *InstructorHandler) Nearby(c *gin.Context) {
	lat, err1 := strconv.ParseFloat(c.Query("lat"), 64)
	lon, err2 := strconv.ParseFloat(c.Query("lon"), 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lon query params are required"})
		return
	}
	radiusKm := 5.0
	if r, err := strconv.ParseFloat(c.Query("radius_km"), 64); err == nil && r > 0 {
		radiusKm = r
	}

	results, err := h.instructors.Nearby(c.Request.Context(), lat, lon, radiusKm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not search nearby instructors"})
		return
	}
	c.JSON(http.StatusOK, results)
}

// ListAll returns every instructor for browsing/booking a future slot,
// regardless of whether they're online right now.
func (h *InstructorHandler) ListAll(c *gin.Context) {
	results, err := h.instructors.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list instructors"})
		return
	}
	c.JSON(http.StatusOK, results)
}
