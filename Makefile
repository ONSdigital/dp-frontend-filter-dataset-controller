BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

export SERVICE_AUTH_TOKEN=POsElEqMI3wc7fk7n8JLzqxWUGyeOpJE6t9H90vzVwQvo1Nin0Fq9hTK6UEjm4rc

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-filter-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	cp rules.json $(BINPATH)

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-filter-dataset-controller -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-filter-dataset-controller

test:
	go test -race -cover ./...

.PHONY: build debug
