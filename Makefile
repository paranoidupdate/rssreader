MAIN_FILE := main.go

# .PHONY tells make these targets don't represent actual files
.PHONY: all build run test lint clean

# Default target
all: lint

# run: Runs the application directly
run:
	go run .

# test: Runs all unit tests with race detection
# test:
#	go test -race -v ./...

# lint: Runs go vet and staticcheck across all packages
# TODO: Add golangci-lint
lint:
	go vet ./...
	staticcheck ./...