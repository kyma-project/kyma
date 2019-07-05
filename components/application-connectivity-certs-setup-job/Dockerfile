FROM golang:1.11-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appconnectivitycertssetupjob .

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job/appconnectivitycertssetupjob .

CMD ["/appconnectivitycertssetupjob"]
