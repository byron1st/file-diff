.PHONY: go-deps
# List outdated direct dependencies
go-deps:
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: check test
check: go-check swift-check

test: go-test swift-test

.PHONY: go-format go-lint go-clean-testcache go-test go-race
# Run linters and modernize checks
go-check: go-format go-lint

# Format and modernize code
go-format:
	@go fmt ./...
	@go fix ./...

# Run linters to check code quality and style
go-lint:
	@go mod tidy
	@golangci-lint run

# Clean test cache to ensure tests run with the latest code changes
go-clean-testcache:
	@go clean -testcache

# Clean test caches and run tests
go-test: go-clean-testcache
	@go test ./...

# Clean test caches, run tests with coverage, and generate an HTML coverage report
go-test-coverage: go-clean-testcache
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | grep total | awk '{print $3}'

# Clean test caches and run race tests
go-race: go-clean-testcache
	@go test -short -race ./...

EXCLUDE := (mocks|docs|cmd)
PACKAGES := $(shell go list ./... | grep -v -E '$(EXCLUDE)')

.PHONY: go-coverage
go-coverage:
	go test $(PACKAGES) -coverprofile=coverage.out
	go tool cover -func=coverage.out | grep total | awk '{print $$3}'

.PHONY: swift-build swift-test swift-check swift-lint swift-format swift-coverage
# Build the Swift package
swift-build:
	@swift build

# Run Swift tests
swift-test:
	@swift test

# Validate the Swift package (build + test)
swift-check: swift-format swift-lint

# Run SwiftLint in strict mode (warnings = errors)
swift-lint:
	@swiftlint lint --config .swiftlint.yml --strict

SWIFT_SOURCES := swift/Sources swift/Tests
swift-format:
	@swiftformat $(SWIFT_SOURCES) --config .swiftformat

BIN_PATH := $(shell swift build --show-bin-path)
XCTEST_BUNDLE := $(shell ls $(BIN_PATH) | grep '\.xctest$$' | head -1)
BUNDLE_NAME := $(shell echo $(XCTEST_BUNDLE) | sed 's/\.xctest//')
TARGET := $(BIN_PATH)/$(XCTEST_BUNDLE)/Contents/MacOS/$(BUNDLE_NAME)
PROFDATA := .build/debug/codecov/default.profdata

swift-coverage:
	@swift test --enable-code-coverage
	@xcrun llvm-cov report "$(TARGET)" -instr-profile="$(PROFDATA)" --ignore-filename-regex=".build|Tests"
