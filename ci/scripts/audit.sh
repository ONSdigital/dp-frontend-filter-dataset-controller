#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-frontend-filter-dataset-controller
  make audit
popd