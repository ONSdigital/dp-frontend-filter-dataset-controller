#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-filter-dataset-controller
  make build && cp build/* $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
