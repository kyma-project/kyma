# NATS Server


## Overview

This Helm chart deploys [NATS](https://nats.io/) using NATS [Helm chart](https://github.com/nats-io/k8s/tree/v0.9.0/helm/charts/nats).



## Installation

To install this chart, follow these steps:

1. Install NATS into the `nats` namespace using Helm 3 :
```bash
kubectl create namespace nats
helm template nats nats -n nats | kubectl apply -f -
```
2. Test the installation:
```bash
$ kubectl -n nats port-forward nats-1 4222
```

## Configuration

### Limits

```yaml
nats:
  # The number of connect attempts against discovered routes.
  connectRetries: 30

  # How many seconds should pass before sending a PING
  # to a client that has no activity.
  pingInterval: 

  # Server settings.
  limits:
    maxConnections: 
    maxSubscriptions: 
    maxControlLine: 
    maxPayload: 

    writeDeadline: 
    maxPending: 
    maxPings: 
    lameDuckDuration: 

  # Number of seconds to wait for client connections to end after the pod termination is requested
  terminationGracePeriodSeconds: 60
```

### Logging

> **NOTE**: It is not recommended to enable trace or debug in production, since enabling it will significantly degrade performance.

```yaml
nats:
  logging:
    debug: 
    trace: 
    logtime: 
    connectErrorReports: 
    reconnectErrorReports: 
```
### Clustering

If clustering is enabled, then a 3-node cluster will be set up. You can find more information in the [NATS documentation](https://docs.nats.io/running-a-nats-service/introduction/running/nats-kubernetes/helm-charts#clustering).

```yaml
cluster:
  enabled: true
  replicas: 3
```
### JetStream

To set up memory and file storage with JetStream, use this configuration: 

```yaml
nats:
  image: nats:alpine

  jetstream:
    enabled: true

    memStorage:
      enabled: true
      size: 2Gi

    fileStorage:
      enabled: true
      size: 1Gi
      storageDirectory: /data/
      storageClassName: default
```
