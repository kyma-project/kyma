.PHONY: test

test:
	@echo '+ Running e2e tests'
ifdef TEST_ARGS
	@go test ./ -args $(TEST_ARGS)
else
	@go test ./
endif

vendor: Gopkg.lock
	@echo '+ Pulling vendored dependencies'
	@dep ensure -v --vendor-only
