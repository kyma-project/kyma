.PHONY: ci-pr ci-master ci-release resolve build-and-test build-image push-image

ci-pr: resolve build-and-test
ci-master: resolve build-and-test
ci-release: resolve build-and-test

resolve:
	dep ensure -v -vendor-only
build-and-test:
	./before-commit.sh ci