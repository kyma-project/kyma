FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-registry

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationregistry ./cmd/applicationregistry

FROM alpine:3.8
LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add curl ngrep

COPY ./docs/api/api.yaml .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-registry/applicationregistry .

CMD ["/applicationregistry"]
