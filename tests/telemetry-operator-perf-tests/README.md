# Telemetry Fluent Bit Perf Tests

## Overview

Small program to deploy a bunch of log pipelines to a Kubernetes cluster. The following pipelines are deployed:
1. Single log pipeline that logs to Loki
2. Multiple log pipelines that log to an HTTP host of choice. The upstream host, port, as well as the number of log pipelines can be set via flags.
In order to simulate upstream unavailability, the `unhealthy-ratio` can be set.

## Usage

Example:
```bash
go run ./... -count=4 -unhealthy-ratio=0.5 -host=mockserver.mockserver -port=1080
```

## Test Setup

Deploy telemetry config file to configure storage path for the filesystem buffer.

```bash
kubectl apply -f deploy/telemetry-config.yaml
```

Install dummy log generator:
```bash
kubectl apply -f deploy/logspammer.yaml
```

Install HTTP mock server (not exposed, only available inside the cluster). The mock server is configured to expose 2 endpoints: /good (returns 201 status code) and /bad (returns 503 status code):
```bash
kubectl apply -f deploy/mockserver.yaml
```

Expose the HTTP mock server via the Kyma Istio Gateway at mockserver.here_comes_my_cluster_hostname (provide your host name):
```bash
cat deploy/mockserver-vs.yaml | sed "s/KYMA_HOST_PLACEHOLDER/here_comes_my_cluster_hostname/g" | kubectl apply -f -
``` 

