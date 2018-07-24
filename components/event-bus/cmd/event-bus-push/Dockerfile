FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/
COPY generated ./generated
COPY internal ./internal
COPY vendor ./vendor
COPY api ./api

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-push/
COPY main.go .
COPY application ./application
RUN ls ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o event-bus-push .

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /root/

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-push/event-bus-push .

EXPOSE 8080

ENTRYPOINT ["/root/event-bus-push"]
