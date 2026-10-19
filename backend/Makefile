-include .env
export

.PHONY: up down run migrate-up migrate-down seed-admin

up:
	docker compose up -d

down:
	docker compose down

run:
	go run ./cmd/api

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

seed-admin:
	go run ./cmd/seedadmin -email=$(EMAIL) -password=$(PASSWORD) -name="$(NAME)"
