.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

test-local: ## Run unit test of integration test engine
	GO111MODULE=on go test -count=1 -v ./pkg/...

resolve-local:
	go mod download
	go mod vendor

compile:
	CGO_ENABLED=0 go build -o main.test ./cmd/main.go

.PHONY: serverless-integration
serverless-integration: ## Run serverless integration scenario with kubectl proxy
	./run-local.sh serverless-integration

.PHONY: git-auth-integration
git-auth-integration: ## Run git authorization integration scenario with kubectl proxy
	./run-local.sh git-auth-integration

.PHONY: simple-tracing
simple-tracing: ## Run simple tracing scenario with kubectl proxy
	./run-local.sh simple-tracing
