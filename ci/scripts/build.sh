#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontend-filter-dataset-controller
  make build && cp build/dp-frontend-filter-dataset-controller $cwd/build
popd
