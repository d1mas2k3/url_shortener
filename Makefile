include .env
export MSYS_NO_PATHCONV=1
export
export PROJECT_ROOT=$(CURDIR)

run-memory:
	@export APP_STORAGE=memory && \
	export LOGGER_FOLDER=${PROJECT_ROOT}/logs && \
	go run ${PROJECT_ROOT}/cmd/server/main.go

run-postgres:
	@export APP_STORAGE=postgres && \
	export LOGGER_FOLDER=${PROJECT_ROOT}/logs && \
	go run ${PROJECT_ROOT}/cmd/server/main.go

test:
	go test ./...

test-race:
	go test -race ./...

up:
	@docker compose --env-file .env up -d url-shortener-postgres

down:
	@docker compose --env-file .env down url-shortener-postgres

init-db:
	@docker exec -i url-shortener-postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} < init.sql

docker-build:
	@docker build -t url-shortener .

docker-run-memory:
	@docker run -p 8080:8080 \
		-e APP_STORAGE=memory \
		-e APP_BASE_URL=${APP_BASE_URL} \
		-e HTTP_ADDR=${HTTP_ADDR} \
		-e HTTP_SHUTDOWN_TIMEOUT=${HTTP_SHUTDOWN_TIMEOUT} \
		-e LOGGER_LEVEL=${LOGGER_LEVEL} \
		-e LOGGER_FOLDER=/tmp/logs \
		url-shortener

test-race-docker:
	@docker run --rm \
		-v ${PROJECT_ROOT}:/app \
		-w /app \
		-e CGO_ENABLED=1 \
		golang:1.25-alpine \
		sh -c "apk add --no-cache gcc musl-dev && go test -race ./..."