# image builder base on golang:1.20.3-alpine3.17
FROM europe-docker.pkg.dev/kyma-project/prod/external/golang@sha256:48c87cd759e3342fcbc4241533337141e7d8457ec33ab9660abe0a4346c30b60 as builder

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

# result container base on alpine:3.17.3
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:b6ca290b6b4cdcca5b3db3ffa338ee0285c11744b4a6abaa9627746ee3291d8d

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client openssl

COPY --from=builder /app /app

ENTRYPOINT ["/app/registry-gc"]
