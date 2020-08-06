FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/console-backend-service
ENV GO111MODULE on

WORKDIR ${BASE_APP_DIR}

#COPY ./go.mod ./go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download

COPY . .

# Run tests
#RUN make verify

# Build app
#RUN make build

# Copy built app
RUN mkdir /app
RUN mv ./console-backend-service /app/main 

COPY ./licenses/ /app/licenses

FROM alpine:3.11
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

# Install certificates
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Copy binary
COPY --from=builder /app /app

# Run app
CMD ["/app/main"]
