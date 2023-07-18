test-local:
	GO111MODULE=on go test -count=1 -v ./pkg/...

resolve-local:
	go mod download
	go mod vendor

compile:
	CGO_ENABLED=0 go build -o main.test ./cmd/main.go

.PHONY: serverless-integration
serverless-integration:
	./run-local.sh serverless-integration

.PHONY: git-auth-integration
git-auth-integration:
	./run-local.sh git-auth-integration

.PHONY: simple-tracing
simple-tracing:
	@echo simple-tracing-not-implememted
