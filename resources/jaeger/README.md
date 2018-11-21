# Jaeger

## Overview
[Jaeger](http://jaeger.readthedocs.io/en/latest/) is a monitoring and tracing tool for microservice-based distributed systems.

## Details
The Envoy sidecar uses Jaeger to trace the request flow in the Istio service mesh. Istio and Envoy use the Zipkin protocol to communicate with the tracing back-end. Jaeger provides compatibility with the Zipkin protocol. This allows you to use Zipkin protocol and clients in Istio, Envoy, and the Kyma services.

For more details, see the [Istio Distributed Tracing](https://istio.io/docs/tasks/telemetry/distributed-tracing.html) documentation.

## Installing Jaeger locally
Jaeger installation is optional, and you cannot install it locally by default. However, you can install it on a Kyma instance and run it locally using Helm.

To install Jaeger, go to the `~/go/src/github.com/kyma-project/kyma/resources/` directory and run the following command:
```bash
$ helm install -n jaeger -f jaeger/values.yaml --namespace kyma-system --set-string global.domainName=kyma.local --set-string global.isLocalEnv=true jaeger/
```
