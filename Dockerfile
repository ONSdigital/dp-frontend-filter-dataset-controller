FROM ubuntu:16.04

WORKDIR /app/

COPY ./build/dp-frontend-filter-dataset-controller .

ENTRYPOINT ./dp-frontend-filter-dataset-controller
