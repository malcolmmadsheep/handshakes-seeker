.DEFAULT_GOAL:=help

docker_volume = handshakes_db
create_db_container = create_db_container

##@ Run

.PHONY: run-dev
run-dev: ## start full application as docker compose
	docker-compose up

.PHONY: build-and-run
build-and-run: build-compose run-dev

##@ Database

.PHONY: setup-db
setup-db:
	docker volume create $(docker_volume)
	docker container stop $(create_db_container) || true
	docker container run \
	--volume $(docker_volume):/var/lib/postgresql/data \
	--rm \
	-d \
	--name $(create_db_container) \
	-e POSTGRES_PASSWORD=password \
	postgres:14
	docker container exec -it $(create_db_container) psql -h localhost -U postgres -c 'create database handshakes;' || true
	docker container stop $(create_db_container)	

.PHONY: add-migration
add-migration: ## create new db migration. Migration name should be provided with "MIGRATION_NAME". Usage example: MIGRATION_NAME=test make add-migration
	if [[ "$(which migrate)" == "" ]]; then go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; fi
	if [[ "${MIGRATION_NAME}" == "" ]]; then echo "MIGRATION_NAME should be provided" && exit 1; fi
	migrate create -ext sql -dir migrations ${MIGRATION_NAME}

.PHONY: clear-db
clear-db:
	docker-compose up -d db
	docker container exec -it handshakes-seeker_db_1 psql -h localhost -U postgres -d handshakes -c 'delete from tasks_queue;' || true
	docker container exec -it handshakes-seeker_db_1 psql -h localhost -U postgres -d handshakes -c 'delete from paths;' || true
	docker-compose down

##@ Build

.PHONY: build-seeker
build-seeker: ## build seeker Docker image
	docker image build -f ./cmd/seeker/Dockerfile --tag seeker .

.PHONY: build-run-seeker
build-run-seeker: build-seeker
	docker container run seeker

.PHONY: build-compose
build-compose:
	docker-compose build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
