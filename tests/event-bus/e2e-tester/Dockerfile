FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/tests/event-bus/
COPY vendor ./vendor

WORKDIR /go/src/github.com/kyma-project/kyma/tests/event-bus/e2e-tester
COPY e2e-tester/e2e-tester.go .
RUN ls ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o e2e-tester .

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /root/

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/event-bus/e2e-tester/e2e-tester .
COPY licenses/ ./licenses

ENTRYPOINT ["/root/e2e-tester"]
