.PHONY: all fmt lint test build check clean docker-build run

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS = -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

all: check build

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files found: $$(gofmt -l .)" && exit 1)

lint:
	go vet ./...

test:
	go test -v -race ./...

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/pipeforge-service ./cmd/pipeforge-service

check: fmt-check lint test build

clean:
	rm -rf bin/ dist/ *.out coverage.html

docker-build:
	docker build -t pipeforge-service:$(VERSION) .

run: build
	./bin/pipeforge-service
