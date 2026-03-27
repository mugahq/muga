VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build lint test cover vet fmt clean

build:  ## Build the muga binary
	go build $(LDFLAGS) -o bin/muga ./cmd/muga

lint:  ## Run golangci-lint
	golangci-lint run ./...

vet:  ## Run go vet
	go vet ./...

fmt:  ## Check formatting
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

test:  ## Run tests
	go test ./... -v

cover:  ## Run tests with coverage (90% threshold)
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' | \
		awk '{if ($$1 < 90) {print "FAIL: coverage " $$1 "% < 90%"; exit 1} else {print "OK: coverage " $$1 "%"}}'

clean:  ## Remove build artifacts
	rm -rf bin/ coverage.out

help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
