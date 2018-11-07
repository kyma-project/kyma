FROM golang:1.11.1 as builder
WORKDIR /go/src/github.com/kyma-project/kyma/components/k8s-dashboard-proxy
COPY cmd ./cmd
COPY util ./util
RUN ls ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo ./cmd/reverseproxy
FROM scratch
ARG version
ENV APP_VERSION $version
WORKDIR /root/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/k8s-dashboard-proxy/reverseproxy .
EXPOSE 8080
ENTRYPOINT ["/root/reverseproxy"]
