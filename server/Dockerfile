FROM golang:1.23 AS build-stage
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . ./
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /telebot

FROM alpine AS release-stage
WORKDIR /
COPY --from=build-stage /telebot /telebot
COPY --from=build-stage /app/etc/config.yml /config.yml
ENTRYPOINT ["/telebot", "--conf", "./config.yml"]