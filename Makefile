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