# syntax=docker/dockerfile:1.0-experimental
FROM golang:1.15-alpine as builder
ARG VERSION
WORKDIR /src

RUN apk add --no-cache make bash

# Add go.mod and go.sum first to maximize caching
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

# Build our application
RUN make build APP_VERSION=${VERSION}

FROM alpine:3.12
ENTRYPOINT ["/usr/bin/{{ .manifest.Name }}"]

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /src/bin/{{ .manifest.Name }} /usr/bin/{{ .manifest.Name }}
