FROM golang:1.21.1 as build
RUN mkdir -p /app
WORKDIR /app

COPY . .
RUN go mod download


RUN CGO_ENABLED=1 go build -o=/app/bin/crypto_trader -ldflags="-X 'main.version=$(shell date)'" /app
RUN chmod +x /app/bin/crypto_trader
