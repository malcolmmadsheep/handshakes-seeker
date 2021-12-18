.DEFAULT_GOAL:=help

##@ Run

.PHONY: run-dev
run-dev: ## start full application as docker compose
	docker-compose up

##@ Build

.PHONY: build-seeker
build-seeker: ## build seeker Docker image
	docker image build -f ./cmd/seeker/Dockerfile --tag seeker .

.PHONY: build-run-seeker
build-run-seeker: build-seeker
	docker container run seeker

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
