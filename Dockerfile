FROM golang:1.19-alpine3.17 AS builder

RUN apk update && apk add --no-cache \
  build-base
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o kuboreleaser ./cmd/kuboreleaser/main.go

FROM alpine:3.17

RUN apk update && apk add --no-cache \
  bash \
  zsh \
  npm \
  diffutils \
  yq

COPY --from=builder /src/kuboreleaser /usr/bin/kuboreleaser

ENTRYPOINT [ "/usr/bin/kuboreleaser" ]
