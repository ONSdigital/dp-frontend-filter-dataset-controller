#!/bin/bash -eux

pushd dp-frontend-filter-dataset-controller
  make test-component
popd
