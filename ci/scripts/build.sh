#!/bin/bash -eux

export BINPATH=$(pwd)/bin
export GOPATH=$(pwd)/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontend-dataset-controller
  go build -tags 'production' -o $BINPATH/dp-frontend-dataset-controller
popd
