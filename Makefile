.PHONY: test test-unit test-integration test-all db-up db-down db-wait shell mod-download clean

# Default: run unit tests (fast, no database)
test: test-unit

# Run unit tests only (no database dependencies)
test-unit:
	@echo "=== Running Unit Tests (Fast - No DB) ==="
	docker compose run --rm app go test -v

# Run integration tests (with databases)
test-integration: db-up db-wait
	@echo "=== Running Integration Tests (With Databases) ==="
	docker compose run --rm app sh -c "cd integration-tests && go test -v"

# Run all tests (unit + integration)
test-all: db-up db-wait
	@echo "=== Running Unit Tests (Fast - No DB) ==="
	docker compose run --rm app go test -v
	@echo "=== Running Integration Tests (With Databases) ==="
	docker compose run --rm app sh -c "cd integration-tests && go test -v"
	@echo "=== All Tests Completed Successfully! ==="

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

# Download Go modules
mod-download:
	@echo "=== Downloading Go Modules ==="
	docker compose run --rm app go mod download
	@docker compose run --rm app sh -c "cd integration-tests && go mod download"

# Open interactive shell (with databases running)
shell: db-up db-wait
	@echo "=== Opening Interactive Shell ==="
	docker compose run --rm app bash

# Open shell without databases (for quick commands)
shell-no-db:
	@echo "=== Opening Interactive Shell (No DB) ==="
	docker compose run --rm app bash

# Clean up Docker resources
clean:
	@echo "=== Cleaning Up Docker Resources ==="
	docker compose down -v --remove-orphans
