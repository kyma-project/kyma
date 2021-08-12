# Build the controller upgrade-job binary
FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20201009-5d8104e0 as builder
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/eventing-controller
WORKDIR $DOCK_PKG_DIR

COPY go.mod go.mod
COPY go.sum go.sum

COPY api/ api/
COPY cmd/eventing-pre-upgrade-job-1-24 cmd/eventing-pre-upgrade-job-1-24
COPY logger/ logger/
COPY options/ options/
COPY pkg/ pkg/
COPY reconciler/ reconciler/
COPY testing/ testing/
COPY utils/ utils/
COPY upgrade-job/ upgrade-job/

# Build
RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o eventing-pre-upgrade-job-1-24 ./cmd/eventing-pre-upgrade-job-1-24

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git
USER nonroot:nonroot

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/eventing-controller/eventing-pre-upgrade-job-1-24 .

COPY licenses/ /licenses/

ENTRYPOINT ["/eventing-pre-upgrade-job-1-24"]
