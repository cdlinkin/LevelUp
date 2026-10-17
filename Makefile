-include .env
export

export PROJECT_ROOT := $(CURDIR)

.DEFAULT_GOAL := help

.PHONY: help \
	env-up env-down env-cleanup env-port-forward env-port-close \
	migrate-create migrate-up migrate-down migrate-action \
	levelup-run levelup-deploy levelup-undeploy logs-cleanup \
	tidy build vet test \
	swagger-gen compose-check ps

help:
	@echo Environment:
	@echo   env-up             start postgres container
	@echo   env-down           stop postgres container
	@echo   env-port-forward   expose postgres on 127.0.0.1:9998
	@echo   env-port-close     stop port forwarder
	@echo   env-cleanup        stop containers and DELETE database files
	@echo Migrations:
	@echo   migrate-create     create migration: make migrate-create seq=name
	@echo   migrate-up         apply all migrations
	@echo   migrate-down       rollback migrations
	@echo Application:
	@echo   levelup-run        run app locally (go run)
	@echo   levelup-deploy     build and run app in docker
	@echo   levelup-undeploy   stop app container
	@echo   logs-cleanup       delete out/logs
	@echo Development:
	@echo   tidy               go mod tidy
	@echo   build              compile all packages
	@echo   vet                go vet
	@echo   test               run tests
	@echo   swagger-gen        regenerate swagger docs
	@echo   compose-check      validate docker-compose.yaml
	@echo   ps                 list project containers

env-up:
	docker compose up -d todoapp-postgres

env-down:
	docker compose down todoapp-postgres

env-cleanup:
	docker compose down todoapp-postgres port-forwarder
	powershell -NoProfile -Command "if (Test-Path '$(PROJECT_ROOT)/out/pgdata') { Remove-Item -Recurse -Force '$(PROJECT_ROOT)/out/pgdata' }"
	@echo Environment files cleaned

env-port-forward:
	docker compose up -d port-forwarder

env-port-close:
	docker compose down port-forwarder

migrate-create:
ifndef seq
	$(error Missing seq parameter. Example: make migrate-create seq=init)
endif
	docker compose run --rm todoapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@$(MAKE) migrate-action action=up

migrate-down:
	@$(MAKE) migrate-action action=down

migrate-action:
ifndef action
	$(error Missing action parameter. Example: make migrate-action action=up)
endif
	@docker compose run --rm todoapp-postgres-migrate -path /migrations -database postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@todoapp-postgres:5432/$(POSTGRES_DB)?sslmode=disable $(action)

levelup-run:
	set POSTGRES_HOST=localhost&& \
	set LOGGER_FOLDER=./out/logs&& \
	go run ./cmd/levelup

levelup-deploy:
	docker compose up -d --build levelup

levelup-undeploy:
	docker compose down levelup

logs-cleanup:
	powershell -NoProfile -Command "if (Test-Path '$(PROJECT_ROOT)/out/logs') { Remove-Item -Recurse -Force '$(PROJECT_ROOT)/out/logs' }"
	@echo Logs cleaned

tidy:
	go mod tidy

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

swagger-gen:
	docker compose run --rm swagger \
		init \
		-g cmd/levelup/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

compose-check:
	docker compose config --quiet

ps:
	docker compose ps