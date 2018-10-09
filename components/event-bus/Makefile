.PHONY: all vet test

all: vet test

vet:
	go list ./... | grep -v generated | xargs go vet

test:
	go list ./... | grep -v generated | xargs go test -v
