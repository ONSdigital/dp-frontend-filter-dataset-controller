BINPATH ?= build

build:
	go build -tags 'production' -o $(BINPATH)/dp-frontent-filter-dataset-controller

debug:
	go build -tags 'debug' -o $(BINPATH)/dp-frontent-filter-dataset-controller
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-frontent-filter-dataset-controller

test:
	go test -tags 'production' ./...

.PHONY: build debug
