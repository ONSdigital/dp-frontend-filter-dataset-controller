FROM ubuntu:16.04

WORKDIR /app/

COPY ./build/dp-frontent-filter-dataset-controller .

ENTRYPOINT ./dp-frontent-filter-dataset-controller
