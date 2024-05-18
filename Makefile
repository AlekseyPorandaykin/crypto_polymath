HOME_PATH := $(shell pwd)

BIN := "./bin/crypto_polymath"
VERSION :=$(shell date)

prepare:
	go mod download
	@go generate $(shell go list ./... | grep -v ./.go/)
	go mod tidy

build:
	CGO_ENABLED=1 go build -o=$(BIN) -ldflags="-X 'main.version=${VERSION}' -X 'github.com/AlekseyPorandaykin/crypto_polymath/cmd.homeDir=${HOME_PATH}'" .

init:
	go install golang.org/x/tools/cmd/goimports@latest

up:
	./bin/crypto_polymath server

linters:
	go vet .
	gofmt -w .
	goimports -w .
	gci write /app
	gofumpt -l -w /app
	golangci-lint run ./...
	gofmt -s -l $(git ls-files '*.go')


go-fix:
	go mod tidy
	gci write ./
	gofumpt -l -w ./

.PHONY: build run build-img run-img version test lint
