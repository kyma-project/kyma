# Telemetry Fluent Bit Perf Tests

## Overview

Small program to deploy a bunch of log pipelines to a Kubernetes cluster. The following pipelines are deployed:
1. Simple log pipeline that logs to Loki
2. Multiple log pipelines that log to an HTTP host of choice. The upstream host, port, as well as the number of log pipelines can be set via flags.
In order to simulate upstream unavailability, the `unhealthy-ratio` can be set.

## Usage

Example:
```bash
go run ./... -count=4 -unhealthy-ratio=0.5 -host=mockserver.mockserver -port=1080
```

## Test Setup

Install dummy log generator:
```bash
kubectl apply -f deploy/logspammer.yaml
```

Install HTTP mock server (not exposed, only available inside the cluster):
```bash
kubectl apply -f deploy/mockserver.yaml
```

Expose the HTTP mockserver via the Kyma Istio Gateway
```bash
kubectl apply -f deploy/mockserver-vs.yaml
``` 
