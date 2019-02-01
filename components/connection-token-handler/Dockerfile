FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/connection-token-handler

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/manager

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/connection-token-handler/manager .

ENTRYPOINT ["/manager"]
