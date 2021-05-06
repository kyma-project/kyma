FROM eu.gcr.io/kyma-project/external/golang:1.16.3-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job
WORKDIR $DOCK_PKG_DIR

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN echo "1000:x:65534:65534:1000:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appconnectivitycertssetupjob .


FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job/appconnectivitycertssetupjob .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job/licenses ./licenses

COPY --from=builder /etc_passwd /etc/passwd
USER 1000

CMD ["/appconnectivitycertssetupjob"]
