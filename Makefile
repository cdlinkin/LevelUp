include .env
export

export PROJECT_ROOT := $(CURDIR)

env-up:
	docker compose up -d todoapp-postgres

env-down:
	docker compose down todoapp-postgres

env-cleanup:
		docker compose down todoapp-postgres port-forwarder && \
		powershell -Command "Remove-Item -Recurse -Force ${PROJECT_ROOT}out/pgdata" && \
		echo "Файлы окружения очищены"; \

env-port-forward:
	docker compose up -d port-forwarder

env-port-close:
	docker compose down port-forwarder

migrate-create:
ifndef seq
	$(error Отсутствует параметр seq. Пример: make migrate-create seq=init)
endif
	docker compose run --rm todoapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down


migrate-action:
ifndef action
	$(error Отсутствует параметр action. Пример: make migrate-action action=up)
endif
	docker compose run --rm todoapp-postgres-migrate -path /migrations -database postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@todoapp-postgres:5432/$(POSTGRES_DB)?sslmode=disable $(action)

logs-cleanup:
	powershell -NoProfile -Command "if (Test-Path '$(PROJECT_ROOT)/out/logs') { Remove-Item -Recurse -Force '$(PROJECT_ROOT)/out/logs' }"
	@echo logs cleaned

levelup-run:
	set POSTGRES_HOST=localhost&& \
	set LOGGER_FOLDER=./out/logs&& \
	go mod tidy && \
	go run cmd/levelup/main.go

levelup-deploy:
	docker compose up -d --build levelup

ps: 
	docker compose ps

compose-check:
	docker compose config --quiet