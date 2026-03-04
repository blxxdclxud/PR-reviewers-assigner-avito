.PHONY: up down test test-integration test-e2e \
        load-seed load-test-create load-test-merge load-test-stats load-test-reassign

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

# Load testing (requires running service: make up)
BASE_URL ?= http://localhost:8080

load-seed:
	bash tests/load/seed.sh $(BASE_URL)

load-test-create:
	k6 run --env BASE_URL=$(BASE_URL) tests/load/k6/create_pr.js

load-test-merge:
	hey -n 5000 -c 10 -m POST \
		-H "Content-Type: application/json" \
		-d '{"pull_request_id":"load-pr-1"}' \
		$(BASE_URL)/pullRequest/merge

load-test-stats:
	hey -n 5000 -c 10 $(BASE_URL)/stats

load-test-reassign:
	k6 run --env BASE_URL=$(BASE_URL) tests/load/k6/reassign.js
