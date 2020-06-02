FROM golang:1.13.11-alpine3.11 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/console-backend-service

ENV GO111MODULE on
ENV CGO_ENABLED 0


WORKDIR ${BASE_APP_DIR}

COPY ./go.mod .
COPY ./go.sum .

COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./main.go .
COPY ./Makefile .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
RUN go mod vendor

# check imports using makefile with custom ignore patterns
RUN apk add make
RUN go get golang.org/x/tools/cmd/goimports
RUN make check-imports

# Build app
RUN go build -v -o main .

# Run tests
RUN go test ./...

# Copy built app
RUN mkdir /app
RUN mv ./main /app/main 

COPY ./licenses/ /app/licenses

FROM alpine:3.10
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

# Install certificates
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Copy binary
COPY --from=builder /app /app

# Run app
CMD ["/app/main"]
