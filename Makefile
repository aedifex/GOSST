# Variables
APP_NAME := gosst
DOCKER_IMAGE := ghosst.azurecr.io/$(APP_NAME)
TAG ?= latest
BINARY := main_linux
COMMIT_SHA := $(shell git rev-parse --short HEAD)
PKG := ./...

# Default target
.PHONY: all
all: check build

# Format Go code
.PHONY: fmt
fmt:
	go fmt $(PKG)

# Static analysis
.PHONY: vet
vet:
	go vet $(PKG)

# Lint (use golangci-lint if available)
.PHONY: lint
lint:
	@golangci-lint run --timeout=5m || echo "golangci-lint not found, skipping lint"

# Dependency tidy/verify
.PHONY: mod
mod:
	go mod tidy
	go mod verify

# Build the Go binary
.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X 'main.CommitSHA=$(COMMIT_SHA)'" -o $(BINARY) .

# Run the application locally
.PHONY: run
run: build
	./$(BINARY)

# Run tests
.PHONY: test
test:
	go test ./... -v

# Build the Docker image
.PHONY: docker-build
docker-build:
	docker build --platform linux/amd64 -t $(DOCKER_IMAGE):$(TAG) .

# Push the Docker image
.PHONY: docker-push
docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(TAG)

# Clean up local files
.PHONY: clean
clean:
	rm -f $(BINARY)
	docker image prune -f

# One target to rule them all: full check before build
.PHONY: check
check: fmt vet lint mod test

# CI Pipeline Target: everything!
.PHONY: ci
ci: check build docker-build docker-push
