FROM golang:1.18

RUN apt-get update && apt-get install -y build-essential

WORKDIR /go/src/bank-api

COPY go.* .
RUN go mod download && go mod verify

COPY . .
RUN mkdir -p /go/bin/bank-api
RUN go build -v -o /go/bin/bank-api ./...

CMD ["/go/bin/bank-api/bank-api"]
