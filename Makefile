# Test runner — one entry point per layer.
# Each layer is independent; CI or a developer can invoke whichever fits.

.PHONY: test-fast test-int test-e2e test-all cover-html clean-test e2e-up e2e-down

# Fast: unit tests only. Should finish in well under a minute.
test-fast:
	cd server && go test ./...
	cd client && npm run test:unit

# Integration: Go tests gated by the `integration` build tag. Boots ephemeral
# Postgres/Redis/MinIO containers via testcontainers-go.
test-int:
	cd server && go test -tags=integration ./... -count=1

# E2E: Playwright against the test docker-compose stack.
test-e2e: e2e-up
	cd e2e && npx playwright test

# All three layers in sequence.
test-all: test-fast test-int test-e2e

# Boot the E2E stack. Idempotent — running twice is harmless.
e2e-up:
	docker compose -f docker-compose.yml -f docker-compose.test.yml up -d --build

e2e-down:
	docker compose -f docker-compose.yml -f docker-compose.test.yml down

# Generate per-package coverage and an HTML drill-down report.
# Targets are documented in TESTING.md; verify manually against the output.
cover-html:
	cd server && go test -coverprofile=coverage.out ./...
	cd server && go tool cover -func=coverage.out
	cd server && go tool cover -html=coverage.out -o coverage.html
	@echo "open server/coverage.html"

clean-test:
	docker compose -f docker-compose.yml -f docker-compose.test.yml down -v
