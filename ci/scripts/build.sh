#!/bin/bash -eux

export BINPATH=$(pwd)/bin
export GOPATH=$(pwd)/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontent-filter-dataset-controller
  go build -tags 'production' -o $BINPATH/dp-frontent-filter-dataset-controller
popd
