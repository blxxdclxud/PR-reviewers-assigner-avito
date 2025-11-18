.PHONY: up down test test-integration test-e2e

# Create .env from .env.example if not exists
env:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo ".env created. Please update values if needed."; \
	else \
		echo ".env already exists"; \
	fi


# Start services
up: env
	docker compose up --build -d

# Stop services
down:
	docker compose down -v

# Run unit tests
test:
	go test -v ./internal/...

# Run integration tests
test-integration:
	docker compose -f docker-compose.test.yml up --build -d
	go test -v -tags=integration ./internal/...

# Run E2E tests
test-e2e:
	docker compose -f docker-compose.test.yml up --build -d
	go test -v -tags=e2e ./tests/e2e/...
