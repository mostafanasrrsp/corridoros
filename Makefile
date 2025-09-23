# CorridorOS Build System
# Supports multiple languages and components

.PHONY: all build test clean lint docs install

# Default target
all: build

# Build all components
build: build-daemons build-sdk build-cli build-labs

# Daemon builds
build-daemons:
	@echo "Building CorridorOS daemons..."
	cd daemon/corrd && cargo build --release
	cd daemon/memqosd && go build -o memqosd .
	cd daemon/fabmand && go build -o fabmand .
	cd daemon/heliopassd && go build -o heliopassd .
	cd daemon/attestd && go build -o attestd .
	cd daemon/compatd && go build -o compatd .
	cd daemon/metricsd && go build -o metricsd .
	cd daemon/securityd && go build -o securityd .

# SDK builds
build-sdk:
	@echo "Building SDKs..."
	cd sdk/rust && cargo build --release
	cd sdk/go && go build -o libcorridor.so -buildmode=c-shared .
	cd sdk/kotlin && ./gradlew build

# CLI builds
build-cli:
	@echo "Building CLI tools..."
	cd cli && go build -o corridor .
	cd cli && go build -o ffm .

# Labs builds
build-labs:
	@echo "Building CorridorLabs..."
	cd labs/physics-decoder && go build -o physics-decoder .
	cd labs/synchrony-analytics && go build -o synchrony-analytics .
	cd labs/helio-sim && go build -o helio-sim .

# Testing
test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	cd daemon/corrd && cargo test
	cd daemon/memqosd && go test ./...
	cd daemon/fabmand && go test ./...
	cd daemon/heliopassd && go test ./...
	cd daemon/attestd && go test ./...
	cd daemon/compatd && go test ./...
	cd daemon/metricsd && go test ./...
	cd daemon/securityd && go test ./...

test-integration:
	@echo "Running integration tests..."
	cd tests/integration && go test ./...

test-hardware:
	@echo "Running hardware-specific tests..."
	cd tests/hardware && go test ./...

# Linting
lint:
	@echo "Running linters..."
	cd daemon/corrd && cargo clippy
	cd daemon/memqosd && golangci-lint run
	cd daemon/fabmand && golangci-lint run
	cd daemon/heliopassd && golangci-lint run
	cd daemon/attestd && golangci-lint run
	cd daemon/compatd && golangci-lint run
	cd daemon/metricsd && golangci-lint run
	cd daemon/securityd && golangci-lint run

# Documentation
docs:
	@echo "Generating documentation..."
	cd docs && make html
	cd daemon/corrd && cargo doc --no-deps
	cd sdk/rust && cargo doc --no-deps

# Installation
install: build
	@echo "Installing CorridorOS..."
	sudo cp daemon/*/target/release/* /usr/local/bin/
	sudo cp daemon/*/* /usr/local/bin/
	sudo cp cli/* /usr/local/bin/
	sudo cp labs/*/* /usr/local/bin/
	sudo systemctl enable corrd memqosd fabmand heliopassd attestd compatd metricsd securityd

# Clean
clean:
	@echo "Cleaning build artifacts..."
	cd daemon/corrd && cargo clean
	cd daemon/memqosd && go clean
	cd daemon/fabmand && go clean
	cd daemon/heliopassd && go clean
	cd daemon/attestd && go clean
	cd daemon/compatd && go clean
	cd daemon/metricsd && go clean
	cd daemon/securityd && go clean
	cd sdk/rust && cargo clean
	cd sdk/go && go clean
	cd sdk/kotlin && ./gradlew clean
	cd cli && go clean
	cd labs/physics-decoder && go clean
	cd labs/synchrony-analytics && go clean
	cd labs/helio-sim && go clean

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	rustup component add rustfmt clippy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	pip install -r requirements-dev.txt

# Docker support
docker-build:
	@echo "Building Docker images..."
	docker build -t corridoros/corrd -f docker/corrd/Dockerfile .
	docker build -t corridoros/memqosd -f docker/memqosd/Dockerfile .
	docker build -t corridoros/fabmand -f docker/fabmand/Dockerfile .

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build all components"
	@echo "  test         - Run all tests"
	@echo "  lint         - Run linters"
	@echo "  docs         - Generate documentation"
	@echo "  install      - Install CorridorOS"
	@echo "  clean        - Clean build artifacts"
	@echo "  dev-setup    - Set up development environment"
	@echo "  docker-build - Build Docker images"
