# image builder base on golang:1.21.1-alpine3.18
FROM golang@sha256:19600fdcae402165dcdab18cb9649540bde6be7274dedb5d205b2f84029fe909 as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/central-application-connectivity-validator
WORKDIR $DOCK_PKG_DIR

COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:65534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o centralapplicationconnectivityvalidator ./cmd/centralapplicationconnectivityvalidator


FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/central-application-connectivity-validator/centralapplicationconnectivityvalidator .

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

CMD ["/app/centralapplicationconnectivityvalidator"]
