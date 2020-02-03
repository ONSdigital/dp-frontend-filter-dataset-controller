#!/bin/bash -eux

export GOPATH=$(pwd)

pushd $GOPATH/dp-frontend-filter-dataset-controller
  make test
popd
