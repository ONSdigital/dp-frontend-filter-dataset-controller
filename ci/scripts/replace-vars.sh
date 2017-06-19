#!/bin/bash -eux

sed dp-frontent-filter-dataset-controller/appspec.yml dp-frontent-filter-dataset-controller/scripts/codedeploy/* -i \
  -e s/\${CODEDEPLOY_USER}/$CODEDEPLOY_USER/g                                     \
  -e s/^CONFIG_BUCKET=.*/CONFIG_BUCKET=$CONFIGURATION_BUCKET/                     \
  -e s/^ECR_REPOSITORY_URI=.*/ECR_REPOSITORY_URI=$ECR_REPOSITORY_URI/             \
  -e s/^GIT_COMMIT=.*/GIT_COMMIT=$(cat build/revision)/                           \
  -e s/^AWS_REGION=.*/AWS_REGION=$AWS_REGION/

mkdir -p artifacts/scripts/codedeploy

cp dp-frontent-filter-dataset-controller/appspec.yml artifacts/
cp dp-frontent-filter-dataset-controller/scripts/codedeploy/* artifacts/scripts/codedeploy
