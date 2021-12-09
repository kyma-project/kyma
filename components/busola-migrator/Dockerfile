FROM eu.gcr.io/kyma-project/external/golang:1.17.6-alpine3.15 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/components/busola-migrator

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /busola-migrator ./main.go

# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest

USER 1000

ENV APP_STATIC_FILES_DIR=/app/static

COPY --from=builder ./busola-migrator /app/busola-migrator
COPY static /app/static

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/app/busola-migrator"]
