BINPATH ?= build

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-filter-dataset-controller

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-filter-dataset-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-filter-dataset-controller

test:
	go test -cover $(shell go list ./... | grep -v /vendor/) -tags 'production' ./...

.PHONY: build debug
