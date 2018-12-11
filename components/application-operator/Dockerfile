FROM golang:1.10.3 as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-operator

RUN mkdir -p $DOCK_PKG_DIR

COPY ./ $DOCK_PKG_DIR
WORKDIR $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/manager

FROM alpine

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-operator/manager .

# Copy RE chart
COPY ./charts/remote-environments   remote-environments/

CMD ["/manager"]
