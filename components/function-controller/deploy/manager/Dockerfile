# image builder base on golang:1.20.5-alpine3.18
FROM europe-docker.pkg.dev/kyma-project/prod/external/golang@sha256:b036c52b3bcc8e4e31be19a7a902bb9897b2bf18028f40fd306a9778bab5771c as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    LIBGIT2_VERSION=1.1.0-r2

RUN apk add --no-cache gcc libc-dev
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.14/community libgit2-dev=${LIBGIT2_VERSION}

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o manager cmd/manager/main.go \
    && mkdir /app \
    && mv ./manager /app/manager

# get latest CA certs from alpine:3.17.3
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:b6ca290b6b4cdcca5b3db3ffa338ee0285c11744b4a6abaa9627746ee3291d8d as certs

RUN apk add --no-cache ca-certificates

# result container base on alpine:3.17.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:e95676db9e4a4f16f6cc01a8915368f82b018cc07aba951c1bd1db586c081388

ENV LIBGIT2_VERSION=1.1.0-r2

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssl
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.14/community libgit2=${LIBGIT2_VERSION}

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/manager"]

