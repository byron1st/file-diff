.PHONY: deps
# List outdated direct dependencies
deps:
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: format lint clean-testcache test race
# Run linters and modernize checks
check: format lint

# Format and modernize code
format:
	@go fmt ./...
	@go fix ./...

# Run linters to check code quality and style
lint:
	@go mod tidy
	@golangci-lint run

# Clean test cache to ensure tests run with the latest code changes
clean-testcache:
	@go clean -testcache

# Clean test caches and run tests
test: clean-testcache
	@go test ./...

# Clean test caches, run tests with coverage, and generate an HTML coverage report
test-coverage: clean-testcache
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | grep total | awk '{print $3}'

# Clean test caches and run race tests
race: clean-testcache
	@go test -short -race ./...

EXCLUDE := (mocks|docs|cmd)
PACKAGES := $(shell go list ./... | grep -v -E '$(EXCLUDE)')

.PHONY: coverage
coverage:
	go test $(PACKAGES) -coverprofile=coverage.out
	go tool cover -func=coverage.out | grep total | awk '{print $$3}'
