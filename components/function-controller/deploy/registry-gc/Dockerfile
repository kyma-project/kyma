# image builder base on golang:1.21.1-alpine3.18
FROM golang@sha256:0c860c7ceba62231d0f99fb92e9d7c1577f26fea794a12c75756a8f64b146e45 as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 

RUN apk add --no-cache gcc libc-dev

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o registry-gc cmd/registry-gc/main.go \
    && mkdir /app \
    && mv ./registry-gc /app/registry-gc

# result container base on alpine:3.17.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:e95676db9e4a4f16f6cc01a8915368f82b018cc07aba951c1bd1db586c081388

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client openssl

COPY --from=builder /app /app

ENTRYPOINT ["/app/registry-gc"]
