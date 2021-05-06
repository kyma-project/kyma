---
title: Jaeger (Tracing)
type: What is...
---

The microservice architecture differs from the traditional monoliths in many aspects. From the request observability perspective, there are asynchronous boundaries among various different microservices that compose a request flow. Moreover, these microservices can have heterogeneous semantics when it comes to monitoring. A tracing solution that provides a holistic view of the request flow helps you to understand the system and take informed decisions regarding troubleshooting and performance optimization.

Tracing in Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend which serves as the query mechanism for displaying information about traces.

>**CAUTION:** Jaeger is designed for application tracing. Traces should not carry sensitive information, such as passwords or credit card numbers.

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