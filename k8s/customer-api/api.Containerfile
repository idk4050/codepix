FROM docker.io/golang:1.19 AS builder

ARG buildcache
ARG modcache

WORKDIR /go/src/customer-api

COPY go.* .
RUN --mount=type=cache,source=${modcache},target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .
RUN mkdir -p /go/bin/customer-api

RUN --mount=type=cache,source=${buildcache},target=/root/.cache/go-build \
    --mount=type=cache,source=${modcache},target=/go/pkg/mod \
    go build -v -o /go/bin/customer-api ./cmd/main.go

FROM docker.io/golang:1.19 AS runner

RUN useradd --create-home customer-api
USER customer-api
WORKDIR /home/customer-api
COPY --from=builder /go/src/customer-api src
COPY --from=builder /go/bin/customer-api bin

CMD ["./bin/main"]
