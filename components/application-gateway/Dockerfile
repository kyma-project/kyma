FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-gateway
WORKDIR $DOCK_PKG_DIR
ENV GOPROXY=http://athens-proxy.c.sap-kyma-prow.internal:80

COPY go.mod .
COPY go.sum .
# RUN echo "$(date +"%Y/%m/%d %T %Z")  [INFO] download" && \
#     go mod download && \ 
#     echo "$(date +"%Y/%m/%d %T %Z") [INFO] download done"

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:5534:nobody:/:" > /etc_passwd

RUN echo "$(date +"%Y/%m/%d %T %Z")  [INFO] build" &&\
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationgateway ./cmd/applicationgateway &&\
    go clean --modcache && \
    echo "$(date +"%Y/%m/%d %T %Z")  [INFO] build done"


FROM eu.gcr.io/kyma-project/external/alpine:3.12.0 as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-gateway/applicationgateway .
COPY --from=builder /app/licenses /app/licenses

COPY --from=builder /etc_passwd /etc/passwd
USER nobody
CMD ["/applicationgateway"]
