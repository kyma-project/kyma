SHELL=bash

.PHONY: ci-pr ci-master ci-release resolve build-and-test generate-mocks

ci-pr: resolve generate-mocks build-and-test
ci-master: resolve generate-mocks build-and-test
ci-release: resolve generate-mocks build-and-test

resolve:
	dep ensure -v -vendor-only
build-and-test:
	./before-commit.sh ci
generate-mocks:
	rm -rf mocks && mkdir mocks && mkdir failery/fixtures/mocks
	go run ./cmd/failery/failery.go -all -recursive -dir="failery/fixtures"
	go run ./cmd/failery/failery.go -print -dir failery/fixtures -name RequesterVariadic > failery/fixtures/mocks/requester_variadic.go
