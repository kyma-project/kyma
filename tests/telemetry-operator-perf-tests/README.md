# Telemetry Fluent Bit Perf Tests

## Overview

This is a small program to deploy a bunch of log pipelines to a Kubernetes cluster. The following pipelines are deployed:
1. A single log pipeline that logs to Loki.
2. Multiple log pipelines that log to an HTTP host of choice. The upstream host, port, as well as the number of log pipelines can be set via flags.
In order to simulate upstream unavailability, the **unhealthy-ratio** flag can be set.

## Usage

Example:
```bash
go run ./... -count=4 -unhealthy-ratio=0.5 -host=mockserver.mockserver -port=1080
```

## Test Setup

You need two Kyma clusters:
- The load-generating cluster (with min/max 4 nodes)
- The http log server sink (with minimum 6 nodes)

## Set up the load-generating cluster

1. Deploy a telemetry config file to configure a storage path for the filesystem buffer:

```bash
kubectl apply -f deploy/telemetry-config.yaml
```

Install a dummy log generator:
```bash
kubectl apply -f deploy/logspammer.yaml
```

## Set up the http log server sink

1. Install an HTTP mock server, which is not exposed, only available inside the cluster.
   The mock server is configured to expose two endpoints: `/good` (returns the status code `201`) and `/bad` (returns the status code `503`):
```bash
kubectl apply -f deploy/mockserver.yaml
```

Expose the HTTP mock server via the Kyma Istio Gateway at `mockserver.{YOUR_CLUSTER_HOSTNAME}` (provide your host name):
```bash
cat deploy/mockserver-vs.yaml | sed "s/KYMA_HOST_PLACEHOLDER/{YOUR_CLUSTER_HOSTNAME}/g" | kubectl apply -f -
``` 

