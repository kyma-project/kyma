.PHONY: generate build build-dev

generate:
	go generate ./...

build:
	go build ./cmd/server

build-dev: generate	build

watch:
	compileDaemon -build="make build" -command="./server"
