FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/
COPY vendor           ./vendor
COPY api              ./api
COPY pkg              ./pkg
COPY internal/trace   ./internal/trace
COPY internal/knative ./internal/knative
COPY internal/ea      ./internal/ea

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/
COPY main.go     .
COPY application ./application
COPY handlers    ./handlers
COPY httpserver  ./httpserver
COPY metrics     ./metrics
COPY publisher   ./publisher
COPY validators  ./validators

RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o event-bus-publish-knative .

FROM alpine:3.7
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /root/
RUN apk --no-cache upgrade && apk --no-cache add curl

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/event-bus-publish-knative .

EXPOSE 8080

ENTRYPOINT ["/root/event-bus-publish-knative"]
