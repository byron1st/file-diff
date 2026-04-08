.PHONY: deps
# List outdated direct dependencies
deps:
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: mockgen
# Generate mocks for interfaces in the project
mockgen:
	command -v mockery >/dev/null 2>&1 || go install github.com/vektra/mockery/v3@latest
	rm -rf internal/mocks && mkdir internal/mocks
	mockery
	find internal/mocks -type f -name '*.go' -exec perl -pi -e 's/interface\{\}/any/g' {} +
	find internal/mocks -type f -name '*.go' -exec gofmt -w {} +

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

# Clean test caches and run race tests
race: clean-testcache
	@go test -short -race ./...
