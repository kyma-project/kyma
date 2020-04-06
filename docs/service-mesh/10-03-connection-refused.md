---
title: Connection refused errors
type: Troubleshooting
---

Mutual TLS (mTLS) is enabled in the Service Mesh by default. As a result, every element of the Service Mesh must have an Istio sidecar with a valid TLS certificate to allow communication. Connection between a service with a sidecar and a service without a sidecar results in a `Connection reset by peer` or a `GOAWAY` response. 

- To enable sidecar injection for Pods of existing Services, restart them after upgrading to Kyma 1.0 or higher.
- If you don't want to use a sidecar, you can whitelist a Service and disable TLS traffic for it by creating a [DestinationRule](https://istio.io/docs/reference/config/networking/destination-rule/).  

>**TIP:** You can use the [`istioctl authn tls-check`](https://istio.io/docs/reference/commands/istioctl/#istioctl-authn-tls-check) command to find out where communication fails.
