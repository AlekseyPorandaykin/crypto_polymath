FROM golang:1.23 AS build

RUN mkdir -p /app
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o=/app/bin/crypto_trader /app

RUN chmod +x /app/bin/crypto_trader
ENTRYPOINT ["/app/bin/crypto_trader"]
