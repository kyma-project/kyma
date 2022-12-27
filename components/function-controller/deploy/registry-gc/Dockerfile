# image builder base on golang:1.19.2-alpine3.16
FROM eu.gcr.io/kyma-project/external/golang@sha256:45d14bdb069fd5e1fee50160a3e15752038f0fabf339a5ed7f342b84701ed3d6 as builder

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

# result container base on alpine:3.16.2
FROM eu.gcr.io/kyma-project/external/alpine@sha256:1304f174557314a7ed9eddb4eab12fed12cb0cd9809e4c28f29af86979a3c870

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client openssl

COPY --from=builder /app /app

ENTRYPOINT ["/app/registry-gc"]
