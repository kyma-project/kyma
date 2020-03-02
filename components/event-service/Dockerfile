FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219 as builder

ARG DOCK_PKG_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/event-service

RUN mkdir -p $DOCK_PKG_DIR

COPY ./ ${DOCK_PKG_DIR}
COPY ./licenses/ /app/licenses

WORKDIR ${DOCK_PKG_DIR}/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o eventservice ./cmd/eventservice

FROM scratch

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /workspace/go/src/github.com/kyma-project/kyma/components/event-service/eventservice .
COPY --from=builder /app/licenses /app/licenses

CMD ["/eventservice"]
