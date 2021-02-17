# Build the controller binary
FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200831-e46c648b as builder
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/eventing-controller
WORKDIR $DOCK_PKG_DIR

COPY go.mod go.mod
COPY go.sum go.sum

COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o eventing-controller main.go

FROM scratch
LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/eventing-controller/eventing-controller .
COPY licenses/ /licenses/
USER nonroot:nonroot

ENTRYPOINT ["/eventing-controller"]
