# Connection Refused Errors

## Symptom

You get either the `Connection reset by peer` response or the `GOAWAY` response when you attempt to establish the connection between a service without a sidecar and a service with a sidecar.

## Cause

By default, mutual TLS (mTLS) is enabled in the service mesh. As a result, every element of the service mesh must have an Istio sidecar with a valid TLS certificate to allow communication.

## Solution

- To add a service without a sidecar to the allowlist and disable mTLS traffic for it, create a DestinationRule resource. See [DestinationRule](https://istio.io/docs/reference/config/networking/destination-rule/).
- To allow connections between a service without a sidecar and a service with a sidecar, create a PeerAuthentication resource in the `PERMISSIVE` mode. See [Peer Authentication](https://istio.io/latest/docs/reference/config/security/peer_authentication/).