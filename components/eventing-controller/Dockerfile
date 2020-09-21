# Build the manager binary
FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200831-e46c648b as builder
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/eventing-controller

WORKDIR $DOCK_PKG_DIR
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go

#COPY api/ api/
#COPY controllers/ controllers/
COPY licenses/ /licenses/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o eventing-controller main.go
RUN ls -lrt
# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM scratch
LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/eventing-controller/eventing-controller .
COPY licenses/ /licenses/
USER nonroot:nonroot

ENTRYPOINT ["/eventing-controller"]
