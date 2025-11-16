.PHONY: up down test test-integration test-e2e

# Start services
up:
	docker compose up --build -d

# Stop services
down:
	docker compose down -v

# Run unit tests
test:
	go test -v ./internal/...

# Run integration tests
test-integration:
	go test -v -tags=integration ./internal/...

# Run E2E tests
test-e2e:
	go test -v -tags=e2e ./tests/e2e/...
