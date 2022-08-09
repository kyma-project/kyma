---
title: Connection refused errors
---

## Symptom

You get a `Connection reset by peer` or a `GOAWAY` response when you attempt to establish a connection between a service without a sidecar and a service with a sidecar.

## Cause

By default, mutual TLS (mTLS) is enabled in the Service Mesh. As a result, every element of the Service Mesh must have an Istio sidecar with a valid TLS certificate to allow communication.

## Remedy

- To whitelist a service without a sidecar and disable mTLS traffic for it, create a [DestinationRule](https://istio.io/docs/reference/config/networking/destination-rule/).
- To allow connections between a service without a sidecar and a service with a sidecar, create a [Peer Authentication](https://istio.io/latest/docs/reference/config/security/peer_authentication/) in the `PERMISSIVE` mode.
