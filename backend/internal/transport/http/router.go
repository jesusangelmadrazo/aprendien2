package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/service"
	"github.com/jmadrazo/clase-drive-backend/internal/transport/http/middleware"
)

type Handlers struct {
	Auth         *AuthHandler
	Instructors  *InstructorHandler
	Availability *AvailabilityHandler
	Lessons      *LessonHandler
	Payments     *PaymentHandler
	Exams        *ExamHandler
	Payouts      *PayoutHandler
	Users        *UserHandler
	Documents    *DocumentHandler
	WS           *WSHandler
}

func NewRouter(auth *service.AuthService, h Handlers, frontendBaseURL, uploadsDir string) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendBaseURL},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.Static("/uploads", uploadsDir)

	api := r.Group("/api")

	authGroup := api.Group("/auth")
	authGroup.POST("/register", h.Auth.Register)
	authGroup.POST("/login", h.Auth.Login)

	instructors := api.Group("/instructors")
	instructors.Use(middleware.AuthRequired(auth))
	{
		instructors.GET("", h.Instructors.ListAll)
		instructors.GET("/nearby", h.Instructors.Nearby)
		instructors.GET("/slots", h.Availability.SlotsForDate)
		instructors.GET("/:id/slots", h.Availability.Slots)
		instructors.POST("/:id/book", middleware.RequireRole("student"), h.Availability.BookSlot)

		me := instructors.Group("/me")
		me.Use(middleware.RequireRole("instructor"))
		me.POST("/profile", h.Instructors.CreateProfile)
		me.GET("/profile", h.Instructors.GetMyProfile)
		me.PATCH("/status", h.Instructors.UpdateStatus)
		me.POST("/location", h.Instructors.UpdateLocation)
		me.POST("/availability", h.Availability.CreateBlock)
		me.GET("/availability", h.Availability.ListMyBlocks)
		me.DELETE("/availability/:id", h.Availability.DeleteBlock)
		me.POST("/schedule-overrides", h.Availability.SetOverride)
		me.GET("/schedule-overrides", h.Availability.ListMyOverrides)
		me.DELETE("/schedule-overrides/:date", h.Availability.DeleteOverride)
		me.GET("/wallet", h.Payouts.GetMyWallet)
		me.GET("/payout-method", h.Payouts.GetMyPayoutMethod)
		me.PUT("/payout-method", h.Payouts.SetMyPayoutMethod)
		me.GET("/documents", h.Documents.GetMyStatus)
		me.POST("/documents", h.Documents.Submit)
	}

	lessons := api.Group("/lessons")
	lessons.Use(middleware.AuthRequired(auth))
	{
		lessons.POST("", middleware.RequireRole("student"), h.Lessons.Create)
		lessons.GET("", h.Lessons.List)
		lessons.GET("/:id", h.Lessons.Get)
		lessons.POST("/:id/start", middleware.RequireRole("instructor"), h.Lessons.Start)
		lessons.POST("/:id/complete", middleware.RequireRole("instructor"), h.Lessons.Complete)
		lessons.POST("/:id/cancel", h.Lessons.Cancel)
		lessons.POST("/offers/:offer_id/respond", middleware.RequireRole("instructor"), h.Lessons.RespondOffer)
	}

	payments := api.Group("/payments")
	payments.Use(middleware.AuthRequired(auth))
	{
		payments.GET("/me", middleware.RequireRole("student"), h.Payments.ListMine)
	}

	exams := api.Group("/exams")
	exams.Use(middleware.AuthRequired(auth))
	{
		exams.GET("/me", middleware.RequireRole("student"), h.Exams.GetMyStatus)
		exams.POST("/submit", middleware.RequireRole("student"), h.Exams.Submit)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(auth), middleware.RequireRole("admin"))
	{
		admin.GET("/exams/pending", h.Exams.ListPending)
		admin.POST("/exams/:id/review", h.Exams.Review)
		admin.GET("/payouts/pending", h.Payouts.ListPendingBalances)
		admin.POST("/payouts/:instructorId", h.Payouts.CreatePayout)
		admin.GET("/documents/pending", h.Documents.ListPending)
		admin.POST("/documents/:id/review", h.Documents.Review)
	}

	api.GET("/users/:id", middleware.AuthRequired(auth), h.Users.GetPublicInfo)

	r.GET("/ws", middleware.AuthRequired(auth), h.WS.Serve)

	return r
}
