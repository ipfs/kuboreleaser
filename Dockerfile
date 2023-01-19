FROM golang:1.19-alpine3.17 AS builder

RUN apk update && apk add --no-cache \
  build-base
ADD . /src
WORKDIR /src
RUN go build -o kuboreleaser ./cmd/kuboreleaser/main.go

FROM alpine:3.17

RUN apk update && apk add --no-cache \
  bash \
  zsh \
  npm \
  diffutils

COPY --from=builder /src/kuboreleaser /usr/bin/kuboreleaser

ENTRYPOINT [ "/usr/bin/kuboreleaser" ]
