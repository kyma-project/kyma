.PHONY: all resolve test test-report validate vet

all: resolve test validate vet

resolve:
	dep ensure -vendor-only -v

test:
	go list ./... | grep -v generated | xargs go test -v

test-report:
	2>&1 go test -v ./... | go2xunit -fail -output unit-tests.xml

validate:
	gometalinter --skip=generated --vendor --deadline=2m --disable-all --enable=vet ./...

vet:
	go list ./... | grep -v generated | xargs go vet
