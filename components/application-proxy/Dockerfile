FROM golang:1.9-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-proxy

RUN mkdir -p $DOCK_PKG_DIR

COPY ./ $DOCK_PKG_DIR
WORKDIR $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationproxy ./cmd/applicationproxy

FROM alpine
RUN apk update && apk add curl ngrep

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-proxy .

CMD ["/applicationproxy"]
