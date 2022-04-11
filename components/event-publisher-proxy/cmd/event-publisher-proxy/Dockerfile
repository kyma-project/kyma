FROM eu.gcr.io/kyma-project/external/golang:1.18.0-alpine3.15 as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/event-publisher-proxy

ARG DOCK_PKG_DIR2=/go/src/github.com/kyma-project/kyma/common

COPY ../../common/. $DOCK_PKG_DIR2

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN apk update

RUN apk add --update alpine-sdk

RUN echo "Running make test"

RUN make test

RUN echo "Completed make test"

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -o event-publisher-proxy ./cmd/event-publisher-proxy

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git
USER nonroot:nonroot

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-publisher-proxy/event-publisher-proxy .

COPY licenses/ /licenses/

ENTRYPOINT ["/event-publisher-proxy"]
