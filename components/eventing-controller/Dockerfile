# Build the controller binary
FROM eu.gcr.io/kyma-project/external/golang:1.18.1-alpine3.15 as builder
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/eventing-controller
WORKDIR $DOCK_PKG_DIR

COPY go.mod go.mod
COPY go.sum go.sum

COPY api/ api/
COPY cmd/ cmd/
COPY logger/ logger/
COPY options/ options/
COPY pkg/ pkg/
COPY controllers/ controllers/
COPY testing/ testing/
COPY utils/ utils/

# Build
RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o eventing-controller ./cmd/eventing-controller

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/eventing-controller/eventing-controller .
COPY licenses/ /licenses/
USER nonroot:nonroot

ENTRYPOINT ["/eventing-controller"]
