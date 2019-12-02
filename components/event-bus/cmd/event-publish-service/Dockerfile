FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/
COPY vendor           ./vendor
COPY api              ./api

COPY internal/trace   ./internal/trace
COPY internal/knative ./internal/knative
COPY internal/ea      ./internal/ea
COPY licenses         ./licenses

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/
COPY main.go     .
COPY application ./application
COPY handlers    ./handlers
COPY httpserver  ./httpserver
COPY metrics     ./metrics
COPY publisher   ./publisher
COPY validators  ./validators
COPY util        ./util

RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o event-publish-service .

FROM alpine:3.7
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /root/
RUN apk --no-cache upgrade && apk --no-cache add curl

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/event-publish-service .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/licenses ./licenses

EXPOSE 8080

ENTRYPOINT ["/root/event-publish-service"]
