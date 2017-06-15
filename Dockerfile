FROM ubuntu:16.04

WORKDIR /app/

COPY ./build/dp-frontend-dataset-controller .

ENTRYPOINT ./dp-frontend-dataset-controller
