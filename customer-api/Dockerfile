FROM golang:1.18

RUN apt-get update && apt-get install -y build-essential

WORKDIR /go/src/customer-api

COPY go.* .
RUN go mod download && go mod verify

COPY . .
RUN mkdir -p /go/bin/customer-api
RUN go build -v -o /go/bin/customer-api ./...

CMD ["/go/bin/customer-api/customer-api"]
