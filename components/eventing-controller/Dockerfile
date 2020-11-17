# Build the controller binary
FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200831-e46c648b as builder
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/eventing-controller
WORKDIR $DOCK_PKG_DIR

COPY go.mod go.mod
COPY go.sum go.sum

COPY api/ api/
COPY pkg/ pkg/
COPY reconciler/ reconciler/
COPY testing/ testing/
COPY utils/ utils/
COPY main.go main.go

# Build
RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o eventing-controller main.go

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/eventing-controller/eventing-controller .
COPY licenses/ /licenses/
USER 1001

ENTRYPOINT ["/eventing-controller"]
