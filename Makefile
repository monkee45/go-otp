# ---- Config ----
DB_NAME := otp_dev
DB_USER := michaelwalsh
DB_PASSWORD := postgres
DB_HOST := localhost
DB_PORT := 5432
DB_SSLMODE := disable

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR := migrations
SEEDS_DIR := seeds

# Goose driver (important)
GOOSE_DRIVER := postgres
GOOSE_DBSTRING := $(DB_URL)

# ---- Commands ----
.PHONY: setup db-create db-drop migrate-up migrate-down migrate-status migrate-reset seed run build new-migration

## Full setup
setup: db-create
# setup: db-create migrate-up seed

## Create database
db-create:
	createdb $(DB_NAME) || true

## Drop database
db-drop:
	dropdb $(DB_NAME) || true

## Run all migrations
migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

## Roll back last migration
migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

## Check migration status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

## Reset database
migrate-reset: db-drop db-create migrate-up

## Seed data
db-seed:
	psql $(DB_URL) -f $(SEEDS_DIR)/*.sql || true

## Run app
run:
	go run main.go config.go
## go run cmd/app/main.go

## Build app
build:
	go build -o bin/app cmd/app/main.go

## Create a new migration
new-migration:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql