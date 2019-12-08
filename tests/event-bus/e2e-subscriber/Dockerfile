FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/tests/event-bus/
COPY vendor ./vendor

WORKDIR /go/src/github.com/kyma-project/kyma/tests/event-bus/e2e-subscriber/
COPY e2e-subscriber/e2e-subscriber.go .
RUN ls ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o e2e-subscriber .

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /root/

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/event-bus/e2e-subscriber/e2e-subscriber .
COPY licenses/ ./licenses

EXPOSE 9000

ENTRYPOINT ["/root/e2e-subscriber"]
