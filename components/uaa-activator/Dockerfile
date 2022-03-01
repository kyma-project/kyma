FROM eu.gcr.io/kyma-project/external/golang:1.17.7-alpine3.15 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/uaa-activator

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /uaa-activator ./main.go

FROM scratch

COPY --from=builder ./uaa-activator /root/uaa-activator

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/uaa-activator"]
