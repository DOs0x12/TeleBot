FROM golang:1.23 AS build-stage
WORKDIR /app
COPY . ./
WORKDIR /app/cmd
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /telebot

FROM alpine AS release-stage
WORKDIR /
COPY --from=build-stage /telebot /telebot
ARG CONFIG
COPY ${CONFIG} /config.yml

ENTRYPOINT ./telebot --conf ./config.yml