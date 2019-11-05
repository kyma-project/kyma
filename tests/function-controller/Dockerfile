FROM golang:1.13.1-alpine as builder

RUN apk add --no-cache ca-certificates

WORKDIR /go/src/github.com/kyma-project/kyma/tests/function-controller
COPY . ./

RUN CGO_ENABLED=0 go test -c ./

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder \
	/go/src/github.com/kyma-project/kyma/tests/function-controller/licenses \
	/go/src/github.com/kyma-project/kyma/tests/function-controller/e2e.test \
	/go/src/github.com/kyma-project/kyma/tests/function-controller/e2e.yaml \
	./

ENTRYPOINT ["./e2e.test"]
