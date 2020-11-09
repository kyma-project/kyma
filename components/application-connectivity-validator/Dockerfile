FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-connectivity-validator
WORKDIR $DOCK_PKG_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./licenses/ /app/licenses
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationconnectivityvalidator ./cmd/applicationconnectivityvalidator

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-connectivity-validator/applicationconnectivityvalidator .
COPY --from=builder /app/licenses /app/licenses
COPY --from=builder /etc/passwd /etc/passwd

CMD ["/applicationconnectivityvalidator"]
