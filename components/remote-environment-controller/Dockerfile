# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/remote-environment-controller

RUN mkdir -p $DOCK_PKG_DIR

COPY ./ $DOCK_PKG_DIR
WORKDIR $DOCK_PKG_DIR

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/manager

# TODO: use alpine if possible
# Copy the controller-manager into a thin image
FROM alpine

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/remote-environment-controller/manager .

# Copy RE chart
COPY ./charts/remote-environments   remote-environments/

CMD ["/manager"]
