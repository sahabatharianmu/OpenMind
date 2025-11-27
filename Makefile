# ==============================================================================
# OS DETECTION & CONFIG
# This block detects the operating system and sets the correct Air config file.
# ==============================================================================
ifeq ($(OS), Windows_NT)
    AIR_CONFIG := .air.windows.toml
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S), Linux)
        AIR_CONFIG := .air.linux.toml
    else ifeq ($(UNAME_S), Darwin)
        AIR_CONFIG := .air.linux.toml
    else
        $(error OS not supported by this Makefile)
    endif
endif

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

.PHONY: run lint lint-check swagger mock_create test

run:
	@echo "Running Air with config: $(AIR_CONFIG)"
	@air -c $(AIR_CONFIG)

lint:
	@golangci-lint run --fix

lint-check:
	@golangci-lint run

swagger:
	@swag init -g ./cmd/server/main.go -o ./pkg/embed/docs

mock_create:
	@mockgen -source="$(SOURCE)" -destination="$(DEST)" -package=mocks

test:
	@go test -v -cover ./...

# ==============================================================================
# MIGRATIONS
# ==============================================================================

.PHONY: migration_up migration_down migration_create

force ?= false
number ?= 1

migration_up:
	@echo "Running migrations up..."
ifeq ($(force), true)
	migrate -path ./pkg/embed/migrations -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable up force $(number)
else
	migrate -path ./pkg/embed/migrations -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable up
endif

migration_down:
	@echo "Running migrations down..."
	migrate -path ./pkg/embed/migrations \
		-database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable \
		down $(number)

migration_goto:
	@echo "Migrating to specific version $(version)..."
	migrate -path ./pkg/embed/migrations \
		-database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable \
		goto $(version)

migration_create:
	@echo "Creating migration file..."
	migrate create -ext sql -dir ./pkg/embed/migrations -seq $(table)