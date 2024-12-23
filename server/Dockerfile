FROM golang:1.23 AS build-stage
WORKDIR /app
COPY . ./
WORKDIR /app/cmd
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /telebot

FROM alpine AS release-stage
WORKDIR /
COPY --from=build-stage /telebot /telebot
COPY --from=build-stage /app/etc/config.yml /config.yml

ARG STORAGE_HOST
ARG STORAGE_DB
ARG STORAGE_USER
ARG STORAGE_PASSWORD

ENV STORAGE_HOST=${STORAGE_HOST}
ENV STORAGE_DB=${STORAGE_DB}
ENV STORAGE_USER=${STORAGE_USER}
ENV STORAGE_PASSWORD=${STORAGE_PASSWORD}

ENTRYPOINT ./telebot --conf ./config.yml --stAddr ${STORAGE_HOST} --stDB ${STORAGE_DB} --stUser ${STORAGE_USER} --stPass ${STORAGE_PASSWORD}