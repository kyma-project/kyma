FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/integration/logging
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /test-logging

FROM alpine:3.10

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.15.2/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

# Variables used for labeling created images
LABEL source=git@github.com:kyma-project/kyma.git

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY --from=builder /test-logging /test-logging
COPY ./licenses/ /licenses
COPY testCounterPod.yaml /

ENTRYPOINT ["/test-logging"]
