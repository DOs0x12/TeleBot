FROM golang:1.23 AS build-stage
WORKDIR /app
COPY . ./
WORKDIR /app/cmd
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /greeter

FROM alpine AS release-stage
RUN apk add --no-cache tzdata
WORKDIR /
COPY --from=build-stage /greeter /greeter
ENTRYPOINT ["/greeter"]