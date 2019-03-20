# Build the manager binary
FROM golang:1.11.5-alpine3.8  as builder

# Copy in the go src
WORKDIR /go/src/github.com/kyma-project/kyma/components/asset-store-controller-manager
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kyma-project/kyma/components/asset-store-controller-manager/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.8
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binary
#


COPY --from=builder /go/src/github.com/kyma-project/kyma/components/asset-store-controller-manager/manager .

#
# Run app
#

ENTRYPOINT ["/manager"]
