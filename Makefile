up:
	docker-compose --file="./docker-compose.yaml" up -d app-external-v1 app-loader app-calculator

up-mac:
	docker compose --file="./docker-compose-dev.yaml" up -d app-external-v1 app-loader app-calculator
up-dev:
	docker-compose  --file="./docker-compose-dev.yaml" up -d app-external-v1 app-loader app-calculator cadvisor rabbit-mq postgres

down:
	docker-compose --file="./docker-compose.yaml" down

ps:
	docker-compose --file="./docker-compose.yaml"  ps

ps-mac:
	docker compose --file="./docker-compose.yaml"  ps

ps-dev:
	docker-compose --file="./docker-compose.yaml"  ps

ps-dev-mac:
	docker compose --file="./docker-compose-dev.yaml"  ps

recreate:
	docker-compose --file="./docker-compose.yaml"  rm -f
	docker-compose --file="./docker-compose.yaml"  pull
	docker-compose --file="./docker-compose.yaml"  up --build -d
	docker-compose --file="./docker-compose.yaml" up --build -d


up-mac:
	docker compose --file="./docker-compose.yaml"  up -d app-external-v1 app-loader app-calculator cadvisor

up-dev-mac:
	docker compose --file="./docker-compose-dev.yaml" up -d cadvisor rabbit-mq postgres

down-mac:
	docker compose --file="./docker-compose.yaml" down

down-dev-mac:
	docker compose --file="./docker-compose-dev.yaml" logs down

ps-mac:
	docker compose --file="./docker-compose.yaml"   ps

recreate-mac:
	docker compose --file="./docker-compose.yaml"  rm -f
	docker compose --file="./docker-compose.yaml"  pull
	docker compose --file="./docker-compose.yaml"  up --build -d
	docker compose --file="./docker-compose.yaml" up --build -d

recreate-dev-mac:
	docker compose --file="./docker-compose-dev.yaml"  rm -f
	docker compose --file="./docker-compose-dev.yaml"  pull
	docker compose --file="./docker-compose-dev.yaml"  up --build -d
	docker compose --file="./docker-compose-dev.yaml" up --build -d

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



.PHONY: up down  recreate go-fix go-linters
