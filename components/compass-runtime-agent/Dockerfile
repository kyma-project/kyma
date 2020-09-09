FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ENV GO111MODULE=on

RUN apt-get install -y ca-certificates

ARG DOCK_PKG_DIR=/compass-runtime-agent
WORKDIR $DOCK_PKG_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o compass-runtime-agent ./cmd

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder $DOCK_PKG_DIR/compass-runtime-agent .
COPY --from=builder /app/licenses /app/licenses

CMD ["/compass-runtime-agent"]