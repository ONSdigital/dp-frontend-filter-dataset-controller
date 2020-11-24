BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	nancy go.sum

.PHONY: build
build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-filter-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	cp rules.json $(BINPATH)

.PHONY: debug
debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-filter-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-filter-dataset-controller

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: build debug
