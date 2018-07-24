FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/
COPY internal ./internal
COPY vendor ./vendor
COPY api ./api

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/
COPY main.go .
COPY controllers ./controllers
COPY handlers ./handlers
COPY application ./application
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o event-bus-publish .

FROM alpine:3.7
LABEL source=git@github.com:kyma-project/kyma.git

ARG version
ENV APP_VERSION $version

WORKDIR /root/
RUN apk --no-cache upgrade && apk --no-cache add curl

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/event-bus-publish .

EXPOSE 8080

ENTRYPOINT ["/root/event-bus-publish"]
