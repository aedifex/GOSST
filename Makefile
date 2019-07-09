# a simple makefile to build both mac & linux binaries

# grab all go files?
GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')

BINARY = "httpGo"
# LOCAL = "httpGO.go"
LINUX_BINARY = "httpGO_Linux"
LDFLAGS=-ldflags "-X main.build_id=${CIRCLE_SHA1} -X main.build_time=${BUILD_TIME}"
LOCAL_LDFLAGS=-ldflags "-X main.build_id=$(shell git rev-parse --short HEAD) -X main.build_time=${BUILD_TIME}"
BUILD_TIME=`date +%Y%m%d%H%M%S`

default: build-local

# Produce a binary for local env
build-local: $(GOFILES)
	go build ${LOCAL_LDFLAGS} -o ${BINARY}

build-linux: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY}

build-docker-image:
	docker build -t chriscircleci/httpgo:${CIRCLE_SHA1} .

# We're building a linux binary because
# our docker image runs Alpine Linux
build-local-docker-image: build-linux
	docker build -t chris/httpgo:coolest .
