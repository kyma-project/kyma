FROM golang:1.15-alpine as builder

ENV BASE_APP_DIR=/go/src/github.com/kyma-project/kyma/components/function-controller \
    GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

COPY ./go.mod .
COPY ./go.sum .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

#
# copy files whitelisted in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o manager cmd/manager/main.go \
    && mkdir /app \
    && mv ./manager /app/manager \
    && if [ -f ${BASE_APP_DIR}/licenses ]; then mv ${BASE_APP_DIR}/licenses /app/licenses; fi

# get latest CA certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# result container
FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/manager"]
