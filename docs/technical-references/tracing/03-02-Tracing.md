---
title: Tracing
type: Details
---

## Overview

[Jaeger](https://www.jaegertracing.io/) is a monitoring and tracing tool for microservice-based distributed systems. Its features include the following:

- Distributed context propagation
- Distributed transaction monitoring
- Root cause analysis
- Service dependency analysis
- Performance and latency optimization

## Usage

The Envoy sidecar uses Jaeger to trace the request flow in the Istio Service Mesh. Jaeger is compatible with the Zipkin protocol, which Istio and Envoy use to communicate with the tracing backend. This allows you to use the Zipkin protocol and clients in Istio, Envoy, and Kyma services.

For details, see [Istio's Distributed Tracing](https://istio.io/docs/tasks/observability/distributed-tracing/).

## Install Tracing locally

Read the [configuration document](/root/kyma#configuration-custom-component-installation-add-a-component) to learn how to install Jaeger locally.

## Access Jaeger

Access the Jaeger UI either locally at `https://jaeger.kyma.local` or on a cluster at `https://jaeger.{domain-of-kyma-cluster}`.
