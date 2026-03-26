.DEFAULT_GOAL := help

COVERAGE_THRESHOLD ?= 70.0

.PHONY: help
help: ## Показать список доступных команд
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: install-deps
install-deps: ## Установить зависимости (mockgen, goose, swag, golangci-lint)
	- go install go.uber.org/mock/mockgen@latest
	- go install github.com/pressly/goose/v3/cmd/goose@latest
	- go install github.com/swaggo/swag/cmd/swag@latest
	- go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: generate
generate: ## Сгенерировать моки и Swagger
	go generate ./...
	swag init -g internal/server/http/router.go -o docs

.PHONY: fmt
fmt: ## Форматировать код
	- go fmt ./...

.PHONY: up
up: ## Запустить все сервисы через Docker Compose
	- docker compose -f deployments/docker-compose.yaml up --build -d

.PHONY: down
down: ## Остановить и удалить контейнеры Docker Compose
	- docker compose -f deployments/docker-compose.yaml down -v

.PHONY: logs
logs: ## Показать логи Docker Compose
	- docker compose -f deployments/docker-compose.yaml logs -f

.PHONY: run-server
run-server: ## Запустить сервер локально
	- go run ./cmd/keeperd

.PHONY: run-client
run-client: ## Запустить клиент локально
	- go run ./cmd/keeper

.PHONY: test
test: ## Запустить все тесты с общим coverage gate
	@packages=$$(go list ./... | awk '!/\/docs$$/ && !/\/internal\/core\/ports\/mocks$$/ && !/\/testhelpers$$/ && !/\/migrations$$/'); \
	coverpkg=$$(printf "%s\n" "$$packages" | paste -sd, -); \
	go test -coverpkg "$$coverpkg" -coverprofile=coverage.out $$packages; \
	total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	awk "BEGIN {exit !($$total >= $(COVERAGE_THRESHOLD))}" || (echo "coverage $$total% is below $(COVERAGE_THRESHOLD)%"; exit 1); \
	echo "Total coverage: $$total%"

.PHONY: test-integration
test-integration: ## Запустить интеграционные тесты с testcontainers (требуется Docker)
	go test -tags=integration -v -count=1 ./internal/repository/...

.PHONY: build-server
build-server: ## Собрать бинарник сервера
	go build -o bin/keeperd ./cmd/keeperd

BUILD_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.1")
BUILD_DATE ?= $(shell date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: build-client
build-client: ## Собрать бинарник клиента
	go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE)" -o bin/keeper ./cmd/keeper

.PHONY: build-client-all
build-client-all: ## Собрать клиент для Linux, macOS и Windows
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE)" -o bin/keeper-linux-amd64 ./cmd/keeper
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE)" -o bin/keeper-darwin-amd64 ./cmd/keeper
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE)" -o bin/keeper-darwin-arm64 ./cmd/keeper
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE)" -o bin/keeper-windows-amd64.exe ./cmd/keeper

.PHONY: lint
lint: ## Запустить линтер
	golangci-lint run

# --- Примеры CLI-команд для ручного тестирования ---

KEEPER ?= go run ./cmd/keeper
SERVER ?= http://localhost:8080

.PHONY: demo-register
demo-register: ## Зарегистрировать тестового пользователя
	$(KEEPER) --server $(SERVER) register -l demo -p demo1234

.PHONY: demo-login
demo-login: ## Войти тестовым пользователем
	$(KEEPER) --server $(SERVER) login -l demo -p demo1234

.PHONY: demo-add-login
demo-add-login: ## Создать секрет типа login
	$(KEEPER) --server $(SERVER) add -t login -d '{"login":"admin","password":"s3cret"}' -m "корпоративный портал"

.PHONY: demo-add-text
demo-add-text: ## Создать секрет типа text
	$(KEEPER) --server $(SERVER) add -t text -d "заметка с приватной информацией" -m "личное"

.PHONY: demo-add-card
demo-add-card: ## Создать секрет типа card
	$(KEEPER) --server $(SERVER) add -t card -d '{"number":"4111111111111111","exp":"12/28","cvv":"123","holder":"IVAN IVANOV"}' -m "основная карта"

.PHONY: demo-add-binary
demo-add-binary: ## Создать бинарный секрет (файл README.md как пример)
	$(KEEPER) --server $(SERVER) add-binary -f README.md -m "копия README"

.PHONY: demo-list
demo-list: ## Показать список секретов
	$(KEEPER) --server $(SERVER) list

.PHONY: demo-get
demo-get: ## Получить секрет по ID (использование: make demo-get id=<id>)
ifndef id
	$(error Использование: make demo-get id=<id>)
endif
	$(KEEPER) --server $(SERVER) get $(id)

.PHONY: demo-update
demo-update: ## Обновить секрет по ID (использование: make demo-update id=<id>)
ifndef id
	$(error Использование: make demo-update id=<id>)
endif
	$(KEEPER) --server $(SERVER) update $(id) -d "обновлённые данные" -m "обновлённые метаданные"

.PHONY: demo-delete
demo-delete: ## Удалить секрет по ID (использование: make demo-delete id=<id>)
ifndef id
	$(error Использование: make demo-delete id=<id>)
endif
	$(KEEPER) --server $(SERVER) delete $(id)

.PHONY: demo-download-binary
demo-download-binary: ## Скачать бинарный файл по ID (использование: make demo-download-binary id=<id>)
ifndef id
	$(error Использование: make demo-download-binary id=<id>)
endif
	$(KEEPER) --server $(SERVER) download-binary $(id) -o /tmp/keeper-download

.PHONY: demo-sync
demo-sync: ## Синхронизировать секреты с сервером
	$(KEEPER) --server $(SERVER) sync

.PHONY: demo-version
demo-version: ## Показать версию клиента
	$(KEEPER) version

.PHONY: demo-logout
demo-logout: ## Выйти из учетной записи
	$(KEEPER) --server $(SERVER) logout

.PHONY: migration
migration: ## Создать новую миграцию (использование: make migration name=имя_миграции)
ifndef name
	$(error Использование: make migration name=<имя_миграции>)
endif
	goose -dir migrations create $(name) go
