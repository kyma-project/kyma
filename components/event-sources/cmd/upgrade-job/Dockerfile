FROM golang:1.15-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/event-sources

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -o upgradejob ./cmd/upgrade-job

FROM gcr.io/distroless/static:nonroot
LABEL source = git@github.com:kyma-project/kyma.git
USER nonroot:nonroot

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-sources/upgradejob /

COPY licenses/ /licenses/

ENTRYPOINT ["/upgradejob"]
