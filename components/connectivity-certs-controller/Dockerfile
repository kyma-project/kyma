FROM golang:1.13-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/connectivity-certs-controller

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR
COPY ./licenses/ /app/licenses

RUN apk add -U --no-cache ca-certificates
RUN CGO_ENABLED=0 GOOS=linux go build -a -o connectivitycertscontroller ./cmd/connectivitycertscontroller

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/connectivity-certs-controller .
COPY --from=builder /app/licenses /app/licenses

CMD ["/connectivitycertscontroller"]
