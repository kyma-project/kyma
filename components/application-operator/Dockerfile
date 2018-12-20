FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-operator

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/manager

FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

COPY charts/application application/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-operator/manager .

CMD ["/manager"]
