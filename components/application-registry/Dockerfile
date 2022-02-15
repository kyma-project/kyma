FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

RUN apk add -U --no-cache ca-certificates

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-registry
WORKDIR $DOCK_PKG_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationregistry ./cmd/applicationregistry

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY ./docs/api/api.yaml .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-registry/applicationregistry .
COPY --from=builder /app/licenses /app/licenses

CMD ["/applicationregistry"]
