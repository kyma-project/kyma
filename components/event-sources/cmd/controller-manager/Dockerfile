FROM eu.gcr.io/kyma-project/external/golang:1.16.0-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/event-sources

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/controller-manager

FROM scratch
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-sources/controller-manager /
COPY licenses/ /licenses/
ENTRYPOINT ["/controller-manager"]
