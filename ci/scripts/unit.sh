#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-filter-dataset-controller
  make test
popd
