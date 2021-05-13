FROM eu.gcr.io/kyma-project/external/golang:1.16.0-alpine as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o manager cmd/manager/main.go \
    && mkdir /app \
    && mv ./manager /app/manager

# get latest CA certs
FROM alpine:3.13.5 as certs
RUN apk --update add ca-certificates

# result container
FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/manager"]
