.PHONY: setup dev build

setup:
	@echo "Checking for go..."
	@command -v go >/dev/null 2>&1 || { echo "go is not installed. Aborting."; exit 1; }
	@echo "Checking for bun..."
	@command -v bun >/dev/null 2>&1 || { echo "bun is not installed. Aborting."; exit 1; }
	@echo "Checking for air..."
	@command -v air >/dev/null 2>&1 || { echo "air is not installed. Please install it by running \"go install github.com/air-verse/air@latest\". Aborting."; exit 1; }
	@echo "Running \"go mod tidy\"..."
	@go mod tidy
	@echo "Running \"bun install\"..."
	@bun install

dev:
	@echo "Starting development environment..."
	@bun run dev & air

build:
	@echo "Building project..."
	@bun run build
	@go build -o ./server
