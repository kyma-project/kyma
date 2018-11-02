FROM golang:1.11.1 as builder
WORKDIR /go/src/k8s-dashboard-proxy/
COPY ./reverseproxy/reverseproxy.go .
RUN ls ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo .
FROM scratch
ARG version
ENV APP_VERSION $version
WORKDIR /root/
COPY --from=builder /go/src/k8s-dashboard-proxy/k8s-dashboard-proxy .
EXPOSE 8080
ENTRYPOINT ["/root/k8s-dashboard-proxy"]