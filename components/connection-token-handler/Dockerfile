FROM eu.gcr.io/kyma-project/external/golang:1.16.3-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/connection-token-handler
WORKDIR $DOCK_PKG_DIR

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:5534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/manager


FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/connection-token-handler/manager .
COPY --from=builder /app/licenses /app/licenses

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

ENTRYPOINT ["/manager"]
