FROM eu.gcr.io/kyma-project/external/golang:1.16.4-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/nats-operator

WORKDIR ${DOCK_PKG_DIR}
COPY . ${DOCK_PKG_DIR}

RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -o nats-operator-doctor ./cmd/nats-operator-doctor/main.go

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git
USER nonroot:nonroot

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/nats-operator/nats-operator-doctor .

COPY licenses/ /licenses/

ENTRYPOINT ["nats-operator-doctor"]
