# Build Docker containers
docker-build:
	@echo "Building Docker containers..."
	@docker compose build

# Run Docker containers
docker-up: docker-build
	@echo "Starting Docker containers..."
	@docker compose up -d

# Run migrations (Docker)
docker-migrate-up:
	@echo "Running migrations in Docker..."
	@docker compose run --rm app go run cmd/migrate/main.go up

# Rollback migrations (Docker)
docker-migrate-down:
	@echo "Rolling back migrations in Docker..."
	@docker compose run --rm app go run cmd/migrate/main.go down

# Show help
help:
	@echo "Available commands:"
	@echo "  make docker-build       - Build Docker containers"
	@echo "  make docker-up          - Build and start Docker containers"
	@echo "  make docker-migrate-up  - Run migrations (Docker)"
	@echo "  make docker-migrate-down- Rollback migrations (Docker)"
	@echo "  make help               - Show this help message"

.PHONY: docker-build docker-up docker-migrate-up docker-migrate-down help
