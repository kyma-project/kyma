# Build the manager binary
FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder:v20190208-813daef as builder
#
# Copy in the go src
#
WORKDIR /go/src/github.com/kyma-project/kyma/components/function-controller
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/
COPY licenses/ licenses/
#
# Build
#
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kyma-project/kyma/components/function-controller/cmd/manager
#
# Copy the controller-manager into a thin image
#
FROM alpine:3.8
LABEL source=git@github.com:kyma-project/kyma.git
WORKDIR /
#
# Copy binary
#
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/function-controller/manager .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/function-controller/licenses/ licenses/
#
# Run app
#
ENTRYPOINT ["/manager"]
