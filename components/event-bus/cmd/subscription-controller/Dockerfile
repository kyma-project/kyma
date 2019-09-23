# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus
COPY vendor/ vendor/
COPY api/ api/

COPY internal/ internal/
COPY generated/ generated/
COPY licenses/ licenses/

WORKDIR /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/subscription-controller
COPY cmd/    cmd/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kyma-project/kyma/components/event-bus/cmd/subscription-controller/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.7
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /root/
RUN apk --no-cache upgrade && apk --no-cache add curl

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/cmd/subscription-controller/manager .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-bus/licenses ./licenses

EXPOSE 8080

ENTRYPOINT ["./manager"]
