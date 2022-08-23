ifdef DEBUG
VERBOSE = -v
endif

.PHONY: resolve
resolve:
	go mod tidy

.PHONY: test
test:
	go test -race -timeout 2m $(VERBOSE) ./...

