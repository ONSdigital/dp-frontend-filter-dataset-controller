BINPATH ?= build

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontend-dataset-controller

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontend-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontend-controller

test:
	go test -tags 'production' ./...

.PHONY: build debug
