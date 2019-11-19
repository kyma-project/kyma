FROM golang:1.13-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /e2e ./cmd/runner

FROM alpine:3.10
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /e2e /

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT ["/e2e"]
