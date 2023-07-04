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
  yq \
  go \
  git \
  jq \
  python3 \
  build-base

RUN git config --global gc.auto 0
RUN git config --global protocol.version 2
RUN git config --global core.sshCommand ""

COPY --from=builder /src/kuboreleaser /usr/bin/kuboreleaser

ENTRYPOINT [ "/usr/bin/kuboreleaser" ]
