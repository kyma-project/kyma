---
title: Connection refused errors
---

## Symptom and cause

Mutual TLS (mTLS) is enabled in the Service Mesh by default. As a result, every element of the Service Mesh must have an Istio sidecar with a valid TLS certificate to allow communication. Attempts to establish connection between a service without a sidecar and a service with a sidecar result in a `Connection reset by peer` or a `GOAWAY` response.

## Remedy

- To whitelist a Service without a sidecar and disable mTLS traffic for it, create a [DestinationRule](https://istio.io/docs/reference/config/networking/destination-rule/).
- To allow connections between a Service without a sidecar and a Service with a sidecar, create a [Peer Authentication](https://istio.io/latest/docs/reference/config/security/peer_authentication/) in the `PERMISSIVE` mode.
