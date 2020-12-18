# Build the controller binary
FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20201009-5d8104e0 as builder
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/eventing-controller
WORKDIR $DOCK_PKG_DIR

COPY go.mod go.mod
COPY go.sum go.sum

COPY api/ api/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY reconciler/ reconciler/
COPY testing/ testing/
COPY utils/ utils/

# Build
RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o eventing-controller-nats ./cmd/eventing-controller-nats

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/eventing-controller/eventing-controller-nats .
COPY licenses/ /licenses/
USER nonroot:nonroot

ENTRYPOINT ["/eventing-controller-nats"]
