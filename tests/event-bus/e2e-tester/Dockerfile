FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/tests/event-bus/
COPY vendor ./vendor

WORKDIR /go/src/github.com/kyma-project/kyma/tests/event-bus/e2e-tester
COPY e2e-tester/e2e-tester.go .
RUN ls ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o e2e-tester .

FROM curlimages/curl:7.66.0
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /test/

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/event-bus/e2e-tester/e2e-tester .
COPY licenses/ ./licenses

ENTRYPOINT ["/test/e2e-tester"]
