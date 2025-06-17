up:
	docker-compose --file="./docker-compose.yaml" up -d  postgres postgres_exporter cadvisor
	docker-compose --file="./docker-compose.yaml" up -d --build app-external-v1 app-loader app-calculator
down:
	docker-compose --file="./docker-compose.yaml" down

ps:
	docker-compose --file="./docker-compose.yaml"  ps

recreate:
	docker-compose --file="./docker-compose.yaml"  rm -f
	docker-compose --file="./docker-compose.yaml"  pull
	docker-compose --file="./docker-compose.yaml"  up --build -d
	docker-compose --file="./docker-compose.yaml" up --build -d

go-fix:
	go mod tidy
	gci write ./
	gofumpt -l -w ./

go-linters:
	go vet .
	gofmt -w .
	goimports -w .
	gci write /app
	gofumpt -l -w /app
	golangci-lint run ./...
	gofmt -s -l $(git ls-files '*.go')


prepare:
	go mod download
	@go generate $(shell go list ./... | grep -v ./.go/)
	go mod tidy

.PHONY: up down ps  recreate go-fix go-linters prepare
