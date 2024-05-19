HOME_PATH := $(shell pwd)

DOCKER_DIR="./deployments/docker-compose.yaml"
BIN := "./bin/crypto_polymath"
VERSION :=$(shell date)

prepare:
	go mod download
	@go generate $(shell go list ./... | grep -v ./.go/)
	go mod tidy

build:
	CGO_ENABLED=1 go build -o=./bin/crypto_polymath -ldflags="-X 'main.version=Sat May 18 23:24:22 UTC 2024' -X 'github.com/AlekseyPorandaykin/crypto_polymath/cmd.homeDir=/projects/crypto_polymath'" .

init:
	go install golang.org/x/tools/cmd/goimports@latest

up:
	./bin/crypto_polymath daemon

linters:
	go vet .
	gofmt -w .
	goimports -w .
	gci write /app
	gofumpt -l -w /app
	golangci-lint run ./...
	gofmt -s -l $(git ls-files '*.go')


up-deploy:
	docker-compose --file=$(DOCKER_DIR) up -d

down-deploy:
	docker-compose --file="./deployments/docker-compose.yaml" up --build  prometheus

ps:
	docker-compose --file=$(DOCKER_DIR) ps

recreate-deploy:
	docker-compose --file=$(DOCKER_DIR) rm -f
	docker-compose --file=$(DOCKER_DIR) pull
	docker-compose --file=$(DOCKER_DIR) up --build -d
	docker-compose --file=$(DOCKER_DIR) up --build -d

go-fix:
	go mod tidy
	gci write ./
	gofumpt -l -w ./

.PHONY: build run build-img run-img version test lint
