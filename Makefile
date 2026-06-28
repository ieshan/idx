.PHONY: setup test test-unit test-integration test-all test-docker vet build tidy ci \
        db-up db-down db-wait shell shell-no-db mod-download clean help

# Modules in the project
MODULES := . ./integration-tests

# Default: run unit tests locally (fast, no database, no Docker)
test: test-unit

# === Setup ===

# Initialize Go workspace, sync deps, tidy all modules
setup:
	@echo "=== Setting up development environment ==="
	go version
	@echo "=== Generating Go workspace ==="
	rm -f go.work go.work.sum
	go work init . ./integration-tests
	@echo "=== Syncing workspace dependencies ==="
	go work sync
	@echo "=== Running go mod tidy in all modules ==="
	@for mod in $(MODULES); do \
		echo "  -> tidying $$mod"; \
		(cd $$mod && go mod tidy); \
	done
	@echo "=== Setup complete ==="

# === Local tests (no Docker required) ===

# Run unit tests locally (fast, no database dependencies)
test-unit:
	@echo "=== Running Unit Tests (Local - No DB) ==="
	go test -v -race -count=1 ./...

# Run integration tests locally (requires databases running on localhost)
test-integration-local:
	@echo "=== Running Integration Tests (Local - Requires DBs) ==="
	cd integration-tests && go test -v -race -count=1 ./...

# Run all tests locally (unit + integration, requires DBs for integration)
test-local: test-unit test-integration-local
	@echo "=== All Local Tests Passed ==="

# === Docker tests ===

# Run unit tests in Docker (no database needed)
test-unit-docker:
	@echo "=== Running Unit Tests (Docker - No DB) ==="
	docker compose run --rm app go test -v -race -count=1 ./...

# Run integration tests in Docker (with databases)
test-integration: db-up db-wait
	@echo "=== Running Integration Tests (Docker - With Databases) ==="
	docker compose run --rm app sh -c "cd integration-tests && go test -v -race -count=1 ./..."
	@$(MAKE) --no-print-directory db-down

# Run all tests in Docker (unit + integration)
test-all: db-up db-wait
	@echo "=== Running Unit Tests (Docker - No DB) ==="
	docker compose run --rm app go test -v -race -count=1 ./...
	@echo "=== Running Integration Tests (Docker - With Databases) ==="
	docker compose run --rm app sh -c "cd integration-tests && go test -v -race -count=1 ./..."
	@echo "=== All Tests Completed Successfully! ==="
	@$(MAKE) --no-print-directory db-down

# Run all tests via Docker Compose up (nakusp-style, single command)
test-docker:
	@echo "=== Running All Tests via Docker Compose ==="
	docker compose up --abort-on-container-exit --exit-code-from test
	docker compose down
	@echo "=== Docker Tests Complete ==="

# === Code quality ===

# Run go vet in all modules
vet:
	@echo "=== Running go vet in all modules ==="
	@for mod in $(MODULES); do \
		echo "  -> vetting $$mod"; \
		(cd $$mod && go vet ./...); \
	done
	@echo "=== Vet complete ==="

# Build all packages in all modules (compiles test files too)
build:
	@echo "=== Building all packages ==="
	@for mod in $(MODULES); do \
		echo "  -> building $$mod"; \
		(cd $$mod && go build ./... 2>/dev/null; go test -run=^$$ -count=1 ./...); \
	done
	@echo "=== Build complete ==="

# Run go mod tidy in all modules + workspace sync
tidy:
	@echo "=== Running go mod tidy in all modules ==="
	@for mod in $(MODULES); do \
		echo "  -> tidying $$mod"; \
		(cd $$mod && go mod tidy); \
	done
	go work sync
	@echo "=== Tidy complete ==="

# === CI ===

# Run full CI pipeline locally (vet -> build -> test)
ci: vet build test-unit
	@echo "=== CI pipeline passed ==="

# Run full CI pipeline in Docker (vet -> build -> test-all)
ci-docker: vet build test-all
	@echo "=== Docker CI pipeline passed ==="

# === Database services ===

# Start database services
db-up:
	@echo "=== Starting Database Services ==="
	docker compose up -d postgres mariadb mongo

# Wait for databases to be healthy (retry loop)
db-wait:
	@echo "=== Waiting for Databases to be Ready ==="
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		sleep 2; \
		PG_READY=$$(docker compose exec postgres pg_isready -U postgres > /dev/null 2>&1; echo $$?); \
		MARIA_READY=$$(docker compose exec mariadb mariadb-admin ping -h localhost -u root -ppassword > /dev/null 2>&1; echo $$?); \
		MONGO_READY=$$(docker compose exec mongo mongosh --quiet --eval "db.adminCommand('ping')" -u root -p password --authenticationDatabase admin > /dev/null 2>&1; echo $$?); \
		if [ "$$PG_READY" = "0" ] && [ "$$MARIA_READY" = "0" ] && [ "$$MONGO_READY" = "0" ]; then \
			echo "=== Databases Ready ==="; \
			exit 0; \
		fi; \
		echo "Waiting for databases (attempt $$i/10)..."; \
	done; \
	echo "ERROR: Databases failed to become ready"; \
	exit 1

# Stop all services
db-down:
	@echo "=== Stopping All Services ==="
	docker compose down

# === Shell ===

# Open interactive shell (with databases running)
shell: db-up db-wait
	@echo "=== Opening Interactive Shell ==="
	docker compose run --rm app bash

# Open shell without databases (for quick commands)
shell-no-db:
	@echo "=== Opening Interactive Shell (No DB) ==="
	docker compose run --rm app bash

# === Cleanup ===

# Clean Go caches, remove generated files, and stop Docker resources
clean:
	@echo "=== Cleaning up ==="
	@for mod in $(MODULES); do \
		(cd $$mod && go clean -cache -testcache); \
	done
	rm -f go.work go.work.sum
	@echo "=== Cleaning Docker resources ==="
	docker compose down -v --remove-orphans
	@echo "=== Clean complete ==="

# Download Go modules (locally)
mod-download:
	@echo "=== Downloading Go Modules ==="
	go mod download
	@cd integration-tests && go mod download

# === Help ===

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Setup:"
	@echo "  setup              Initialize Go workspace, sync deps, tidy all modules"
	@echo "  tidy               Run go mod tidy in all modules + workspace sync"
	@echo "  mod-download       Download Go modules for all modules"
	@echo ""
	@echo "Testing (local - no Docker):"
	@echo "  test               Run unit tests locally (default)"
	@echo "  test-unit          Same as 'test' - fast unit tests"
	@echo "  test-integration-local  Run integration tests locally (requires DBs)"
	@echo "  test-local         Run unit + integration tests locally"
	@echo ""
	@echo "Testing (Docker):"
	@echo "  test-unit-docker   Run unit tests in Docker"
	@echo "  test-integration   Run integration tests in Docker (starts DBs)"
	@echo "  test-all           Run unit + integration tests in Docker"
	@echo "  test-docker        Run all tests via docker compose up (single command)"
	@echo ""
	@echo "Code quality:"
	@echo "  vet                Run go vet in all modules"
	@echo "  build              Build all packages in all modules"
	@echo "  ci                 Run vet + build + test (local CI simulation)"
	@echo "  ci-docker          Run vet + build + test-all (Docker CI simulation)"
	@echo ""
	@echo "Database services:"
	@echo "  db-up              Start database services"
	@echo "  db-down             Stop all services"
	@echo "  db-wait            Wait for databases to be ready"
	@echo ""
	@echo "Shell:"
	@echo "  shell              Open interactive shell with DBs running"
	@echo "  shell-no-db        Open shell without databases"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean              Clean Go caches, remove go.work, stop Docker"
