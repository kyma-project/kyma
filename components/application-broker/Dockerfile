FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/components/application-broker
ADD . $SRC_DIR

RUN apk --no-cache add curl

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /application-broker ./cmd/broker

# Final image
FROM scratch

COPY ./licenses/ /root/licenses
COPY --from=builder /usr/bin/curl /bin/curl
COPY --from=builder ./application-broker /root/application-broker

# Variables used for labeling created images
LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/application-broker"]
