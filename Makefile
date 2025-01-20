# Variables
APP_NAME := gosst
DOCKER_IMAGE := ghosst.azurecr.io/$(APP_NAME)
TAG := latest
BINARY := main_linux

# Default target
.PHONY: all
all: build

# Build the Go binary
.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY) .

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

# CI Pipeline Target
.PHONY: ci
ci: build test docker-build docker-push