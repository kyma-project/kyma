# Health check proxy

## Overview

The Health check proxy is designed to check liveness or readiness endpoints in the applications with mTLS enabled which are built from the scratch image.

The Kubernetes API server cannot reach Pods with mTLS enabled via httpGet, because of that it cannot make a direct call to the liveness or readiness probe for status.

In this case, the best way to check the status of our application is to provide a light-weight binary to our application image, which will proxy our application status to the Kubernetes API server.

## Usage

### Build image

To build image locally, use command:

```bash
make build-local
```

### Use as base image

To attach the Health check proxy binary into your scratch image, just build from it in your dockerfile:

```dockerfile
FROM eu.gcr.io/kyma-project/health-check-proxy:TAG
```

The health-check binary is located under `/health-check` path.

### Build image remotely

In case you want to use the Health check proxy binary from your image which isn't built from the scratch, you must build a binary on your own:

```dockerfile
RUN go get github.com/kyma-project/kyma/components/health-check-proxy/...
RUN go build github.com/polskikiel/kyma/components/health-check-proxy/...
```

It builds a `health-check` binary on root directory.

## Parameters

You can configure the Health check proxy binary with the following flags:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **path** | Defines an URL path to endpoint which exposes a status. |  |
| **host** | Defines the Host address. | `localhost` |
| **statusPort** | Specifies port of the status endpoint. |  |
| **retry** | Specifies a number of retries for calling the given endpoint. | `1` |

## Example

See the example on how to use the Health check proxy to check the status of your application:

```yaml
readinessProbe:
  exec:
    command:
      - "/health-check"
      - "-path=ready"
      - "-statusPort={{ .Values.broker.statusPort }}"
livenessProbe:
  exec:
    command:
      - "/health-check"
      - "-path=live"
      - "-statusPort={{ .Values.broker.statusPort }}"
```