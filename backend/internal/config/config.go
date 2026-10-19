package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Port            string `env:"PORT" envDefault:"8080"`
	DatabaseURL     string `env:"DATABASE_URL,required"`
	RedisAddr       string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	JWTSecret       string `env:"JWT_SECRET,required"`
	MPAccessToken   string `env:"MP_ACCESS_TOKEN"`
	MPWebhookURL    string `env:"MP_WEBHOOK_URL"`
	FrontendBaseURL string `env:"FRONTEND_BASE_URL" envDefault:"http://localhost:5173"`
	// PlatformFeePercent is added on top of the instructor's quoted price —
	// the instructor always receives their full rate, the fee is what the
	// platform earns. No real payment provider is wired yet (see
	// internal/provider/manual), so this only affects simulated checkouts
	// for now.
	PlatformFeePercent float64 `env:"PLATFORM_FEE_PERCENT" envDefault:"10"`

	// Exam eligibility thresholds, straight from the user's rules: the
	// theory exam unlocks after N completed lessons; the practice exam
	// unlocks after N total minutes of completed lessons (15h, matching the
	// official IMM requirement already shown on the student's home page).
	ExamTheoryUnlockLessons   int `env:"EXAM_THEORY_UNLOCK_LESSONS" envDefault:"8"`
	ExamPracticeUnlockMinutes int `env:"EXAM_PRACTICE_UNLOCK_MINUTES" envDefault:"900"`

	// UploadsDir is where exam result photos are saved and served from
	// (r.Static("/uploads", ...)). Local disk, not a persistent volume —
	// fine for an MVP, but uploads won't survive a redeploy without one.
	UploadsDir string `env:"UPLOADS_DIR" envDefault:"./uploads"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
