FROM golang:1.13-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /testserver ./cmd/fakelambda

FROM alpine:3.10
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /testserver /

RUN apk update && apk add ca-certificates && apk add curl && rm -rf /var/cache/apk/*

ENTRYPOINT ["/testserver"]
