#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontend-filter-dataset-controller
  make build && cp build/* $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
