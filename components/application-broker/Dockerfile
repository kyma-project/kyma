FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/components/application-broker
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /application-broker ./cmd/broker

FROM alpine:latest as certs
RUN apk --no-cache add ca-certificates

FROM scratch

COPY --from=builder ./application-broker /root/application-broker
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ./licenses/ /root/licenses

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/application-broker"]
