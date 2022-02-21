FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20220217-8d32835a as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/event-publisher-proxy

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -o event-publisher-proxy ./cmd/event-publisher-proxy

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git
USER nonroot:nonroot

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-publisher-proxy/event-publisher-proxy .

COPY licenses/ /licenses/

ENTRYPOINT ["/event-publisher-proxy"]
