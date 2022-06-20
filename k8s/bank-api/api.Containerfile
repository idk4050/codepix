FROM docker.io/golang:1.19 AS builder

ARG buildcache
ARG modcache

WORKDIR /go/src/bank-api

COPY go.* .
RUN --mount=type=cache,source=${modcache},target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .
RUN mkdir -p /go/bin/bank-api

RUN --mount=type=cache,source=${buildcache},target=/root/.cache/go-build \
    --mount=type=cache,source=${modcache},target=/go/pkg/mod \
    go build -v -o /go/bin/bank-api ./cmd/main.go

FROM docker.io/golang:1.19 AS runner

RUN useradd --create-home bank-api
USER bank-api
WORKDIR /home/bank-api
COPY --from=builder /go/src/bank-api src
COPY --from=builder /go/bin/bank-api bin

CMD ["./bin/main"]
