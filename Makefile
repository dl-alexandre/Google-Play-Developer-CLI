# Makefile for gpd - Google Play Developer CLI

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Binary name
BINARY_NAME = gpd
BINARY_DIR = bin

# Build flags
LDFLAGS = -ldflags "-X github.com/dl-alexandre/gpd/pkg/version.Version=$(VERSION) \
	-X github.com/dl-alexandre/gpd/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/dl-alexandre/gpd/pkg/version.BuildTime=$(BUILD_TIME)"

# Optimized build flags (strip debug info, trim paths)
LDFLAGS_OPTIMIZED = -ldflags "-s -w \
	-X github.com/dl-alexandre/gpd/pkg/version.Version=$(VERSION) \
	-X github.com/dl-alexandre/gpd/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/dl-alexandre/gpd/pkg/version.BuildTime=$(BUILD_TIME)"

# Platforms
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-optimized build-small build-optimized-all build-all clean test deps tidy lint install help format install-hooks \
	benchmark benchmark-compare benchmark-regression benchmark-baseline \
	test-unit test-integration test-e2e test-coverage-threshold test-flaky security check vet \
	apidrift apidrift-check apidrift-json apidrift-markdown apidrift-build apidrift-verbose apidrift-multi

# Default target
all: deps build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gpd

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,) ./cmd/gpd; \
		echo "Built $(BINARY_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done

# Build optimized binary (strip debug info, trim paths)
build-optimized:
	@echo "Building optimized $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 $(GOBUILD) -trimpath $(LDFLAGS_OPTIMIZED) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gpd
	@echo "✓ Optimized binary built: $(BINARY_DIR)/$(BINARY_NAME)"
	@echo "  Size: $$(ls -lh $(BINARY_DIR)/$(BINARY_NAME) | awk '{print $$5}')"
	@echo ""
	@echo "Testing binary..."
	@./$(BINARY_DIR)/$(BINARY_NAME) version

# Build small binary (fully optimized, static, compressed)
build-small:
	@echo "Building small $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 $(GOBUILD) -trimpath $(LDFLAGS_OPTIMIZED) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gpd
	@echo "✓ Optimized binary: $$(ls -lh $(BINARY_DIR)/$(BINARY_NAME) | awk '{print $$5}')"
	@if command -v upx >/dev/null 2>&1; then \
		echo "Compressing with UPX..."; \
		upx --best --lzma $(BINARY_DIR)/$(BINARY_NAME) 2>/dev/null || echo "⚠ UPX compression failed (may not be supported)"; \
		echo "✓ Final binary: $$(ls -lh $(BINARY_DIR)/$(BINARY_NAME) | awk '{print $$5}')"; \
	else \
		echo "⚠ upx not installed. Binary not compressed."; \
		echo "  Install: brew install upx (macOS) or apt-get install upx (Linux)"; \
	fi
	@echo ""
	@echo "Testing binary..."
	@./$(BINARY_DIR)/$(BINARY_NAME) version

# Run full optimization script
build-optimized-all:
	@echo "Running full optimization build..."
	@bash ./scripts/build-optimized.sh all

# Run all checks (format, vet, lint, test)
.PHONY: check
check: format vet lint test
	@echo "✓ All checks passed"

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Tidy go modules
tidy:
	@echo "Tidying go modules..."
	$(GOMOD) tidy

# Run tests
# By default, run only unit tests (exclude integration tests for speed)
test:
	@echo "Running unit tests..."
	@bash -o pipefail -c '$(GOTEST) -v -race -cover -tags=unit ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'

# Run all tests including integration (slower)
test-all:
	@echo "Running all tests (unit + integration)..."
	@bash -o pipefail -c '$(GOTEST) -v -race -cover ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'

# Run only unit tests
test-unit:
	@echo "Running unit tests..."
	@bash -o pipefail -c '$(GOTEST) -v -race -cover -tags=unit -count=1 ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'

# Run only integration tests
test-integration:
	@echo "Running integration tests..."
	@bash -o pipefail -c '$(GOTEST) -v -race -tags=integration -count=1 ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'

# Run E2E tests (if any)
test-e2e:
	@echo "Running E2E tests..."
	@bash -o pipefail -c '$(GOTEST) -v -race -tags=e2e -count=1 ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'

# Run tests multiple times to detect flakiness
test-flaky:
	@echo "Running tests 5 times to detect flaky tests..."
	@bash -o pipefail -c '$(GOTEST) -race -count=5 ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'

# Run tests with coverage threshold check (fails if coverage < 40%)
test-coverage-threshold: test-coverage
	@echo "Checking coverage threshold..."
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	threshold=40.0; \
	if (( $$(echo "$$coverage < $$threshold" | bc -l) )); then \
		echo "❌ Coverage $$coverage% is below threshold $$threshold%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$coverage% meets threshold $$threshold%"; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@bash -o pipefail -c '$(GOTEST) -v -race -coverprofile=coverage.out ./... 2>&1 | sed "/malformed LC_DYSYMTAB/d"'
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Generate checksums
checksums:
	@echo "Generating checksums..."
	@cd $(BINARY_DIR) && \
	for file in $(BINARY_NAME)*; do \
		if [ -f "$$file" ]; then \
			shasum -a 256 "$$file" >> checksums.txt; \
		fi \
	done
	@echo "Checksums written to $(BINARY_DIR)/checksums.txt"

# Format code
format:
	@echo "Formatting code..."
	@gofmt -w -s .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Run security scan with gosec
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@gosec -quiet -conf .github/config/.gosec.json ./... || echo "Security scan completed with warnings"

# Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@git config core.hooksPath .githooks
	@echo "Hooks installed from .githooks/"

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

# Development: run the CLI
run:
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gpd
	@./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

# Development: watch and rebuild
watch:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Run: go install github.com/cosmtrek/air@latest"; \
	fi

# Generate completions
completions: build
	@mkdir -p completions
	@./$(BINARY_DIR)/$(BINARY_NAME) config completion bash > completions/$(BINARY_NAME).bash
	@./$(BINARY_DIR)/$(BINARY_NAME) config completion zsh > completions/_$(BINARY_NAME)
	@./$(BINARY_DIR)/$(BINARY_NAME) config completion fish > completions/$(BINARY_NAME).fish
	@echo "Shell completions generated in completions/"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@mkdir -p .artifacts/benchmarks
	@$(GOTEST) -run '^$$' -bench . -benchmem -count=5 ./cmd/... ./internal/... ./pkg/... | tee .artifacts/benchmarks/benchmark-$(shell date +%Y%m%d-%H%M%S).txt
	@echo "Results saved to .artifacts/benchmarks/"

# Compare current benchmarks against main branch
benchmark-compare:
	@echo "Building benchcheck tool..."
	@$(GOBUILD) -o $(BINARY_DIR)/benchcheck ./cmd/benchcheck
	@echo "Saving current branch benchmarks..."
	@mkdir -p .artifacts/benchmarks
	@$(GOTEST) -run '^$$' -bench . -benchmem -count=5 ./cmd/... ./internal/... ./pkg/... | tee .artifacts/benchmarks/current.txt
	@echo "Checking out main branch..."
	@git stash
	@git checkout main
	@echo "Running main branch benchmarks..."
	@$(GOTEST) -run '^$$' -bench . -benchmem -count=5 ./cmd/... ./internal/... ./pkg/... | tee .artifacts/benchmarks/main.txt
	@echo "Restoring working branch..."
	@git checkout -
	@git stash pop
	@echo "\n=== Benchmark Comparison ==="
	@$(BINARY_DIR)/benchcheck --baseline .artifacts/benchmarks/main.txt --current .artifacts/benchmarks/current.txt

# Quick benchmark regression check (use saved baseline)
benchmark-regression:
	@if [ ! -f .artifacts/benchmarks/baseline.txt ]; then \
		echo "No baseline found. Run: make benchmark && cp .artifacts/benchmarks/benchmark-*.txt .artifacts/benchmarks/baseline.txt"; \
		exit 1; \
	fi
	@echo "Building benchcheck tool..."
	@$(GOBUILD) -o $(BINARY_DIR)/benchcheck ./cmd/benchcheck
	@echo "Running current benchmarks..."
	@mkdir -p .artifacts/benchmarks
	@$(GOTEST) -run '^$$' -bench . -benchmem -count=5 ./cmd/... ./internal/... ./pkg/... | tee .artifacts/benchmarks/current.txt
	@echo "\n=== Regression Check ==="
	@$(BINARY_DIR)/benchcheck --baseline .artifacts/benchmarks/baseline.txt --current .artifacts/benchmarks/current.txt

# Save current benchmark as new baseline
benchmark-baseline:
	@mkdir -p .artifacts/benchmarks
	@latest=$$(ls -t .artifacts/benchmarks/benchmark-*.txt 2>/dev/null | head -1); \
	if [ -z "$$latest" ]; then \
		echo "No benchmark files found. Run: make benchmark"; \
		exit 1; \
	fi; \
	cp "$$latest" .artifacts/benchmarks/baseline.txt; \
	echo "Set $$latest as new baseline"

# Help
help:
	@echo "Google Play Developer CLI (gpd) Makefile"
	@echo ""
	@echo "Build Targets:"
	@echo "  all                   - Build the binary (default)"
	@echo "  build                 - Build for current platform"
	@echo "  build-optimized       - Build optimized binary (strip debug, ~20MB)"
	@echo "  build-small           - Build smallest binary (UPX compressed, ~10-15MB)"
	@echo "  build-optimized-all   - Build all optimization levels with size comparison"
	@echo "  build-all             - Build for all platforms"
	@echo ""
	@echo "Test & Quality:"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  lint              - Run linter"
	@echo "  benchmark         - Run all benchmarks"
	@echo "  benchmark-compare - Compare against main branch"
	@echo "  benchmark-regression - Check for regressions vs baseline"
	@echo "  benchmark-baseline   - Save current benchmark as baseline"
	@echo ""
	@echo "Maintenance:"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy go modules"
	@echo "  format       - Format code with gofmt and goimports"
	@echo "  security     - Run security scan with gosec"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  checksums    - Generate SHA256 checksums"
	@echo "  completions  - Generate shell completions"
	@echo "  install-hooks - Install git hooks"
	@echo "  apidrift     - Check for API drift"
	@echo ""
	@echo "Development:"
	@echo "  run          - Build and run (use ARGS=... for arguments)"
	@echo "  watch        - Watch and rebuild (requires 'air')"
	@echo "  version      - Show version info"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make benchmark"
	@echo "  make benchmark-regression"
	@echo "  make run ARGS='version'"
	@echo "  make build-all"
	@echo "  make apidrift"
	@echo "  make apidrift-check"
	@echo ""
	@echo "API Drift Detection:"
	@echo "  apidrift           - Detect API drift (text output)"
	@echo "  apidrift-check     - Check drift with threshold (fails if > 10)"
	@echo "  apidrift-json      - Output drift report as JSON"
	@echo "  apidrift-markdown  - Output drift report as Markdown"
	@echo "  apidrift-verbose   - Run with verbose output"

# API Drift Detection Targets
# Build the apidrift tool
apidrift-build:
	@echo "Building apidrift tool..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/apidrift ./cmd/apidrift

# Run drift detection with text output
apidrift: apidrift-build
	@echo "Running API drift detection..."
	./$(BINARY_DIR)/apidrift

# Run drift detection with verbose output
apidrift-verbose: apidrift-build
	@echo "Running API drift detection (verbose)..."
	./$(BINARY_DIR)/apidrift -verbose

# Run drift detection and fail if score exceeds threshold
apidrift-check: apidrift-build
	@echo "Checking API drift with threshold..."
	./$(BINARY_DIR)/apidrift -threshold 10 -verbose

# Run drift detection and output JSON
apidrift-json: apidrift-build
	@echo "Running API drift detection (JSON output)..."
	@mkdir -p .artifacts/drift
	./$(BINARY_DIR)/apidrift -output .artifacts/drift/report.json -format json
	@echo "Report saved to .artifacts/drift/report.json"

# Run drift detection and output Markdown
apidrift-markdown: apidrift-build
	@echo "Running API drift detection (Markdown output)..."
	@mkdir -p .artifacts/drift
	./$(BINARY_DIR)/apidrift -output .artifacts/drift/report.md -format markdown
	@echo "Report saved to .artifacts/drift/report.md"

# Check multiple Google APIs (similar to CI)
apidrift-multi: apidrift-build
	@echo "Checking multiple Google APIs..."
	@mkdir -p .artifacts/multi-api
	@for api in androidpublisher drive gmail calendar sheets; do \
		echo "Checking $$api..."; \
		url="https://www.googleapis.com/discovery/v1/apis/$$api/v3/rest"; \
		if [ "$$api" = "androidpublisher" ]; then \
			url="https://www.googleapis.com/discovery/v1/apis/$$api/v3/rest"; \
		elif [ "$$api" = "gmail" ]; then \
			url="https://www.googleapis.com/discovery/v1/apis/$$api/v1/rest"; \
		elif [ "$$api" = "calendar" ] || [ "$$api" = "sheets" ]; then \
			url="https://www.googleapis.com/discovery/v1/apis/$$api/v3/rest"; \
		fi; \
		./$(BINARY_DIR)/apidrift -discovery-url "$$url" -output ".artifacts/multi-api/$$api.json" || true; \
	done
	@echo "Multi-API reports saved to .artifacts/multi-api/"
