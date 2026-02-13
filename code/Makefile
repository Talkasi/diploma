COMPOSE_FILE := infra/docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)
DB_SERVICE := db
API_SERVICE := api
APP_HTTP_ADDR_LOCAL ?= :8081
DEV_IMAGE := logistics-dev

-include infra/.env
export

.PHONY: help up down logs logs-api logs-db db-wait migrate-up rebuild-api run-api-local run-importer-local run-importer-container test-local test-container psql-container

help:
	@echo "Доступные команды:"
	@echo "  make up            - поднять БД и API (docker compose)"
	@echo "  make down          - остановить и удалить контейнеры"
	@echo "  make logs          - логи всех сервисов"
	@echo "  make logs-db       - логи БД"
	@echo "  make logs-api      - логи API"
	@echo "  make db-wait       - ждать готовность БД"
	@echo "  make migrate-up    - применить миграцию 00001_init.sql"
	@echo "  make rebuild-api   - пересобрать образ API"
	@echo "  make run-api-local - запуск API через go run (по умолчанию на $(APP_HTTP_ADDR_LOCAL))"
	@echo "  make run-importer-local  - запуск импортера через go run"
	@echo "  make run-importer-container - запуск импортера внутри контейнера API"
	@echo "  make test-local          - запустить тесты Go локально"
	@echo "  make test-container      - запустить тесты Go внутри контейнера"
	@echo "  make psql-container      - открыть psql внутри контейнера БД"

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

logs-api:
	$(COMPOSE) logs -f $(API_SERVICE)

logs-db:
	$(COMPOSE) logs -f $(DB_SERVICE)

db-wait:
	@DB_CONTAINER="$$( $(COMPOSE) ps -q $(DB_SERVICE) )"; \
	if [ -z "$$DB_CONTAINER" ]; then \
		echo "контейнер БД не запущен. выполните 'make up'"; exit 1; fi; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
		if docker exec "$$DB_CONTAINER" pg_isready -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" >/dev/null 2>&1; then \
			echo "БД готова"; exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "БД не стала готовой вовремя"; exit 1

migrate-up: db-wait
	@DB_CONTAINER="$$( $(COMPOSE) ps -q $(DB_SERVICE) )"; \
	docker exec -i "$$DB_CONTAINER" sh -c 'PGPASSWORD="$${POSTGRES_PASSWORD}" psql -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}"' < migrations/00001_init.sql

rebuild-api:
	$(COMPOSE) build $(API_SERVICE)

run-api-local:
	cd backend && APP_HTTP_ADDR=$(APP_HTTP_ADDR_LOCAL) DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) DB_MAX_CONNS=$(DB_MAX_CONNS) go run ./cmd/api

run-importer-local:
	cd backend && DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) DB_MAX_CONNS=$(DB_MAX_CONNS) go run ./cmd/importer

run-importer-container:
	$(COMPOSE) run --rm $(API_SERVICE) ./importer

test-local:
	cd backend && go test ./...

test-container:
	docker build -f backend/Dockerfile --target dev -t $(DEV_IMAGE) backend
	docker run --rm -w /src $(DEV_IMAGE) go test ./...

psql-container: db-wait
	@DB_CONTAINER="$$( $(COMPOSE) ps -q $(DB_SERVICE) )"; \
	docker exec -it "$$DB_CONTAINER" sh -c 'PGPASSWORD="$${POSTGRES_PASSWORD}" psql -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}"'
