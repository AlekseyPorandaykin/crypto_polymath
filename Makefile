up:
	docker compose --file="./docker-compose.yaml" up --build postgres
	docker-compose --file="./docker-compose.yaml" up -d --build  cadvisor
	docker-compose --file="./docker-compose.yaml" up -d --build postgres_exporter
	docker-compose --file="./docker-compose.yaml" up -d --build app-external-v1 app-loader app-calculator
down:
	docker compose --file="./docker-compose.yaml" down

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

# === Testing ===

test:
	go test ./...

test-unit:
	go test -short -count=1 ./core/... ./internal/... ./cmd/... ./pkg/...

test-core:
	go test ./core/...

test-race:
	go test -race -count=3 ./core/... ./internal/... ./cmd/...

test-cover:
	go test -coverprofile=coverage.out ./core/... ./internal/... ./cmd/... ./pkg/...
	go tool cover -func=coverage.out
	@echo "---"
	@echo "HTML report: go tool cover -html=coverage.out -o coverage.html"

test-fuzz:
	go test -fuzz=Fuzz -fuzztime=60s ./core/trading/...

test-fuzz-short:
	go test -fuzz=Fuzz -fuzztime=10s ./core/trading/...

test-bdd:
	go test -v -count=1 ./tests/bdd/...

test-golden-update:
	go test ./core/trading/ -run=Test.*_golden -update

test-acceptance:
	docker compose -f docker-compose-dev.yaml up -d --wait
	go test -tags=acceptance -timeout=120s -v ./tests/acceptance/...

test-smoke:
	go test -tags=smoke -timeout=30s -v ./tests/smoke/...

test-torture:
	go test -tags=torture -timeout=10m -race -v ./tests/torture/...

test-mutation:
	@command -v gremlins >/dev/null 2>&1 || { echo "Install: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest"; exit 1; }
	gremlins unleash ./core/trading/...

bench-core:
	go test ./core/... -bench=. -benchmem -run=^$$

test-all: test-unit test-race test-cover
	@echo "=== All local tests passed ==="

.PHONY: up down ps recreate go-fix go-linters prepare
.PHONY: test test-unit test-core test-race test-cover test-fuzz test-fuzz-short test-bdd
.PHONY: test-golden-update test-acceptance test-smoke test-torture test-mutation
.PHONY: bench-core test-all
