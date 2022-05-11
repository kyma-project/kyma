FROM eu.gcr.io/kyma-project/external/golang:1.18.1-alpine3.15 as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/central-application-gateway
WORKDIR $DOCK_PKG_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:5534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationgateway ./cmd/applicationgateway


FROM eu.gcr.io/kyma-project/external/alpine:3.14.2 as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/central-application-gateway/applicationgateway .
COPY --from=builder /app/licenses /app/licenses

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

CMD ["/app/applicationgateway"]
