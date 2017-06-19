#!/bin/bash -eux

export GOPATH=$(pwd)/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontent-filter-dataset-controller
  go test -tags 'production' ./...
popd
