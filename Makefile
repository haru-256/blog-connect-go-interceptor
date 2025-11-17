.DEFAULT_GOAL := help

.PHONY: init
init: ## Initial setup
	go mod tidy
	mise install
	buf config init

.PHONY: lint
lint:  ## Lint proto and go files
	buf lint
	golangci-lint run --config=./.golangci.yml ./...

.PHONY: fmt
fmt:  ## Format proto and go files
	buf format -w .
	go fmt ./...

.PHONY: test
test: ## Run tests (skip DB tests by default)
	go test ./...

.PHONY: generate-buf
generate:  ## Generate gRPC code
	buf dep update
	buf generate

.PHONY: run-server
run-server: ## Run server
	go run ./cmd/server/main.go

.PHONY: list-server-services
list-server-services: ## List server services via reflection
	buf curl http://localhost:8081 --list-methods --http2-prior-knowledge

.PHONY: get-user
get-user: ## Get user information
	go run ./cmd/client/main.go -op get-user

.PHONY: list-users
list-users: ## List users
	go run ./cmd/client/main.go -op list-users

.PHONY: update-users
update-users: ## Update users
	go run ./cmd/client/main.go -op update-users

.PHONY: chat
chat: ## chat
	go run ./cmd/client/main.go -op chat

# .PHONY: get-user
# get-user: ## Get user information
# 	buf curl --protocol grpc --http2-prior-knowledge \
# 		--schema . \
# 		--data '{"user_id":"1"}' \
# 		http://localhost:8081/myservice.v1.MyService/GetUser

# .PHONY: list-users
# list-users: ## List users
# 	buf curl --protocol grpc --http2-prior-knowledge \
# 		--schema . \
# 		--data '{"page_size":"3", "page_token":""}' \
# 		http://localhost:8081/myservice.v1.MyService/ListUsers

# .PHONY: update-users
# update-users: ## Update users
# 	buf curl --protocol grpc --http2-prior-knowledge \
# 		--schema . \
# 		--data '{"users":[{"user_id":"1", "name":"New Name1"}, {"user_id":"2", "name":"New Name2"}]} {"users":[{"user_id":"3", "name":"New Name3"}]}' \
# 		http://localhost:8081/myservice.v1.MyService/UpdateUsers

.PHONY: up
up: ## up all
	tilt up -d

.PHONY: down
down: ## down all
	tilt down

.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
