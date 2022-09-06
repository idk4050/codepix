FROM docker.io/golang:1.19 AS builder

ARG buildcache
ARG modcache

WORKDIR /go/src/example-bank-api

COPY go.* .
RUN --mount=type=cache,source=${modcache},target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .
RUN mkdir -p /go/bin/example-bank-api

RUN --mount=type=cache,source=${buildcache},target=/root/.cache/go-build \
    --mount=type=cache,source=${modcache},target=/go/pkg/mod \
    go build -v -o /go/bin/example-bank-api ./cmd/main.go

FROM docker.io/golang:1.19 AS runner

RUN useradd --create-home example-bank-api
USER example-bank-api
WORKDIR /home/example-bank-api
COPY --from=builder /go/src/example-bank-api src
COPY --from=builder /go/bin/example-bank-api bin

CMD ["./bin/main"]
