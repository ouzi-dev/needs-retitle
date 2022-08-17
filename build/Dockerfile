ARG GOLANG_VERSION=1.18.5

FROM quay.io/ouzi/go-builder:${GOLANG_VERSION} as builder

ARG VERSION=canary

COPY . /go/github.com/ouzi-dev/needs-retitle

WORKDIR /go/github.com/ouzi-dev/needs-retitle

RUN VERSION=${VERSION} make build test

FROM alpine:3.16

RUN apk --update --no-cache add ca-certificates

RUN adduser -D needs-retitle

USER needs-retitle

COPY --from=builder /go/github.com/ouzi-dev/needs-retitle/bin/needs-retitle /usr/local/bin/needs-retitle

ENTRYPOINT ["/usr/local/bin/needs-retitle"]