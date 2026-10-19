package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jmadrazo/clase-drive-backend/internal/config"
	"github.com/jmadrazo/clase-drive-backend/internal/platform"
	manualprovider "github.com/jmadrazo/clase-drive-backend/internal/provider/manual"
	pgrepo "github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	redisrepo "github.com/jmadrazo/clase-drive-backend/internal/repository/redis"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
	transporthttp "github.com/jmadrazo/clase-drive-backend/internal/transport/http"
	"github.com/jmadrazo/clase-drive-backend/internal/transport/ws"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := platform.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisClient, err := platform.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	userRepo := pgrepo.NewUserRepo(pool)
	instructorRepo := pgrepo.NewInstructorRepo(pool)
	lessonRepo := pgrepo.NewLessonRepo(pool)
	availabilityRepo := pgrepo.NewAvailabilityRepo(pool)
	scheduleOverrideRepo := pgrepo.NewScheduleOverrideRepo(pool)
	paymentRepo := pgrepo.NewPaymentRepo(pool)
	examRepo := pgrepo.NewExamRepo(pool)
	payoutRepo := pgrepo.NewPayoutRepo(pool)
	payoutMethodRepo := pgrepo.NewPayoutMethodRepo(pool)
	documentRepo := pgrepo.NewInstructorDocumentRepo(pool)
	geoRepo := redisrepo.NewGeoRepo(redisClient)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	instructorService := service.NewInstructorService(instructorRepo, userRepo, geoRepo)
	// autoApprove=true: no real gateway is wired yet, so the simulated
	// provider's checkout is treated as settled immediately. Flip this once
	// a real domain.PaymentProvider (e.g. Mercado Pago) replaces it and
	// approval instead comes from its webhook.
	paymentService := service.NewPaymentService(paymentRepo, manualprovider.New(), cfg.PlatformFeePercent, true)
	examService := service.NewExamService(examRepo, lessonRepo, cfg.ExamTheoryUnlockLessons, cfg.ExamPracticeUnlockMinutes)
	payoutService := service.NewPayoutService(paymentRepo, payoutRepo, payoutMethodRepo)
	documentService := service.NewDocumentService(documentRepo, instructorRepo)

	hub := ws.NewHub()
	matcher := service.NewMatcher(lessonRepo, instructorRepo, userRepo, geoRepo, hub)
	availabilityService := service.NewAvailabilityService(availabilityRepo, scheduleOverrideRepo, lessonRepo, instructorRepo, paymentService, hub)

	handlers := transporthttp.Handlers{
		Auth:         transporthttp.NewAuthHandler(authService),
		Instructors:  transporthttp.NewInstructorHandler(instructorService),
		Availability: transporthttp.NewAvailabilityHandler(availabilityService),
		Lessons:      transporthttp.NewLessonHandler(lessonRepo, matcher),
		Payments:     transporthttp.NewPaymentHandler(paymentRepo),
		Exams:        transporthttp.NewExamHandler(examService, userRepo, cfg.UploadsDir),
		Payouts:      transporthttp.NewPayoutHandler(payoutService),
		Users:        transporthttp.NewUserHandler(userRepo),
		Documents:    transporthttp.NewDocumentHandler(documentService, userRepo, cfg.UploadsDir),
		WS:           transporthttp.NewWSHandler(hub, matcher, instructorService),
	}

	router := transporthttp.NewRouter(authService, handlers, cfg.FrontendBaseURL, cfg.UploadsDir)

	logger.Info("starting server", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
