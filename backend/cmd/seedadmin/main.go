package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jmadrazo/clase-drive-backend/internal/config"
	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/platform"
	pgrepo "github.com/jmadrazo/clase-drive-backend/internal/repository/postgres"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

func main() {
	email := flag.String("email", "", "admin email")
	password := flag.String("password", "", "admin password")
	fullName := flag.String("name", "Admin", "admin full name")
	phone := flag.String("phone", "000000000", "admin phone")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Println("usage: seedadmin -email=admin@example.com -password=secret [-name=\"Admin\"] [-phone=...]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := platform.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Println("db error:", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := pgrepo.NewUserRepo(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	u, err := authService.Register(*email, *phone, *password, *fullName, domain.RoleAdmin)
	if err != nil {
		fmt.Println("could not create admin:", err)
		os.Exit(1)
	}

	fmt.Printf("admin created: %s (%s)\n", u.FullName, u.Email)
}
