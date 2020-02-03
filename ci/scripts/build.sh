#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd

pushd $GOPATH/dp-frontend-filter-dataset-controller
  make build && cp build/* $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
