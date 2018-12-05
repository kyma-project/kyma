# Jaeger

## Overview
[Jaeger](http://jaeger.readthedocs.io/en/latest/) is a monitoring and tracing tool for microservice-based distributed systems.

## Details
The Envoy sidecar uses Jaeger to trace the request flow in the Istio service mesh. Istio and Envoy use the Zipkin protocol to communicate with the tracing back-end. Jaeger provides compatibility with the Zipkin protocol. This allows you to use Zipkin protocol and clients in Istio, Envoy, and the Kyma services.


## Installation
While Jager installs automatically during cluster installation, local Jaeger installation is optional. To run Jaeger locally, install it on a Kyma instance and run using Helm.

