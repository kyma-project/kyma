FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-registry

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationregistry ./cmd/applicationregistry

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY ./docs/api/api.yaml .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-registry/applicationregistry .

CMD ["/applicationregistry"]
