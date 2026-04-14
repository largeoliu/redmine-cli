.PHONY: build test clean install lint lint-fix test-coverage install-linter install-hooks e2e fuzz fuzz-client fuzz-output benchmark benchmark-full integration integration-setup integration-cleanup

BINARY_NAME=redmine
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME).exe ./cmd

test:
	go test -v -race ./...

clean:
	rm -rf bin/

run:
	go run ./cmd

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

install-linter:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4

e2e:
	go test -v -count=1 ./test/e2e/...

fuzz:
	go test -fuzz=Fuzz -fuzztime=30s ./internal/client/...
	go test -fuzz=Fuzz -fuzztime=30s ./internal/output/...

fuzz-client:
	go test -fuzz=Fuzz -fuzztime=60s ./internal/client/...

fuzz-output:
	go test -fuzz=Fuzz -fuzztime=60s ./internal/output/...

benchmark:
	go test -bench=. -benchmem ./internal/client/...
	go test -bench=. -benchmem ./internal/output/...

benchmark-full:
	go test -bench=. -benchmem -benchtime=5s ./...

integration:
	@if [ -z "$(REDMINE_URL)" ] || [ -z "$(REDMINE_API_KEY)" ]; then \
		echo "Error: REDMINE_URL and REDMINE_API_KEY must be set"; \
		echo ""; \
		echo "Usage:"; \
		echo "  REDMINE_URL=https://redmine.example.com REDMINE_API_KEY=xxx make integration"; \
		echo ""; \
		echo "Optional: Set REDMINE_PROJECT_ID to isolate tests to a specific project"; \
		echo "  REDMINE_PROJECT_ID=1 REDMINE_URL=... REDMINE_API_KEY=... make integration"; \
		exit 1; \
	fi
	@echo "Running integration tests..."
	@echo "REDMINE_URL=$(REDMINE_URL)"
	@if [ -n "$(REDMINE_PROJECT_ID)" ]; then \
		echo "REDMINE_PROJECT_ID=$(REDMINE_PROJECT_ID) (isolated to project)"; \
	else \
		echo "Warning: REDMINE_PROJECT_ID not set, tests may affect any project"; \
	fi
	go test -v -count=1 ./test/integration/...

integration-setup:
	@if [ ! -f test/integration/.env ]; then \
		cp test/integration/.env.example test/integration/.env; \
		echo "Created test/integration/.env - please edit with your credentials"; \
		echo ""; \
		echo "Important: Set REDMINE_PROJECT_ID to isolate tests to a specific project"; \
	else \
		echo "test/integration/.env already exists"; \
	fi

integration-with-env:
	@if [ -f test/integration/.env ]; then \
		set -a; . test/integration/.env; set +a; \
		go test -v -count=1 ./test/integration/...; \
	else \
		echo "Error: test/integration/.env not found. Run 'make integration-setup' first"; \
		exit 1; \
	fi

integration-cleanup:
	@if [ -z "$(REDMINE_URL)" ] || [ -z "$(REDMINE_API_KEY)" ]; then \
		echo "Error: REDMINE_URL and REDMINE_API_KEY must be set"; \
		exit 1; \
	fi
	@echo "Running cleanup test to remove test issues..."
	go test -v -count=1 -run TestCleanupTestIssues ./test/integration/...

install-hooks:
	@echo "Installing Git hooks..."
	@if [ ! -d .git/hooks ]; then \
		mkdir -p .git/hooks; \
	fi
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed successfully!"
	@echo ""
	@echo "The following hooks are now active:"
	@echo "  - pre-commit: Runs lint checks before each commit"
