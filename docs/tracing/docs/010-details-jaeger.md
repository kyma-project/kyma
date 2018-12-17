---
title: Jaeger
type: Details
---

## Overview
[Jaeger](http://jaeger.readthedocs.io/en/latest/) is a monitoring and tracing tool for microservice-based distributed systems. Its features include the following:

- Distributed context propagation
- Distributed transaction monitoring
- Root cause analysis
- Service dependency analysis
- Performance and latency optimization

## Usage 

The Envoy sidecar uses Jaeger to trace the request flow in the Istio Service Mesh. Jaeger is compatible with the Zipkin protocol, which Istio and Envoy use to communicate with the tracing back end. This allows you to use the Zipkin protocol and clients in Istio, Envoy, and the Kyma services.

For details, see [Istio's Distributed Tracing](https://istio.io/docs/tasks/telemetry/distributed-tracing.html).

## Install Jaeger locally
While Jager installs automatically during cluster installation, local Jaeger installation is optional. You can install Jaeger on a Kyma instance and run it locally using Helm.

To install Jaeger locally, go to the `~/go/src/github.com/kyma-project/kyma/resources/` directory and run the following command:
```bash
$ helm install -n jaeger -f jaeger/values.yaml --namespace kyma-system --set-string global.domainName=kyma.local --set-string global.isLocalEnv=true jaeger/
```
## Access Jaeger

Access the Jaeger UI either locally at `https://jaeger.kyma.local` or on a cluster at `https://jaeger.{domain-of-kyma-cluster}`. 


