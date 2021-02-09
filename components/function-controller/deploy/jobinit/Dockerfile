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

RUN go build -ldflags "-s -w" -a -o jobinit cmd/jobinit/main.go \
&& mkdir /app \
&& mv ./jobinit /app/jobinit \
&& if [ -f ${BASE_APP_DIR}/licenses ]; then mv ${BASE_APP_DIR}/licenses /app/licenses; fi

# result container
FROM alpine:3.13

RUN apk add --no-cache \
  openssh-client \
  ca-certificates

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]