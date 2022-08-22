# image builder base on golang:1.19.0-alpine3.16                                     
FROM golang@sha256:78767a599f5758e3fd60fd065de9dc40aec1ad3b86ded2bddc92640a465cb515 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/busola-migrator

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /busola-migrator ./main.go

# hadolint ignore=DL3007
FROM gcr.io/distroless/static@sha256:35817ff4e3441ebfa430cfca98121e43cbf4a134412285c2460a9464c97ac2ca

USER 1000

ENV APP_STATIC_FILES_DIR=/app/static

COPY --from=builder ./busola-migrator /app/busola-migrator
COPY static /app/static

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/app/busola-migrator"]
