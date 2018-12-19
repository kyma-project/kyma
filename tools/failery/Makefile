SHELL=bash

.PHONY: ci-pr ci-master ci-release resolve build-and-test integration fixture

ci-pr: resolve build-and-test integration
ci-master: resolve build-and-test integration
ci-release: resolve build-and-test integration

resolve:
	dep ensure -v -vendor-only
build-and-test:
	./before-commit.sh ci

# Original targets
integration:
	rm -rf mocks
	go run ./cmd/failery/failery.go -all -recursive -cpuprofile="failery.prof" -dir="failery/fixtures"
fixture:
	go run ./cmd/failery/failery.go -print -dir failery/fixtures -name RequesterVariadic > failery/fixtures/mocks/requester_variadic.go
