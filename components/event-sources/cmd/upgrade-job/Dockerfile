FROM golang:1.14-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/event-sources

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/http-adapter

FROM scratch
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-sources/upgrade-job /
COPY licenses/ /licenses/
ENTRYPOINT ["/upgrade-job"]
