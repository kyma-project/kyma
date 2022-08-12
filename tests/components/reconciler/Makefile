ifdef DEBUG
VERBOSE = -v
endif

.PHONY: resolve
resolve:
	go mod tidy -compat=1.17

.PHONY: test
test:
	go test -race -timeout 2m $(VERBOSE) ./...

