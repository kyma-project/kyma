---
title: Connection refused errors after upgrade
type: Troubleshooting
---

Mutual TLS (mTLS) is enabled in the Service Mesh by default. As a result, every element of the Service Mesh must have an Istio sidecar to allow TLS communication. If you don't want to use a sidecar, you can whitelist a service and disable TLS traffic for it by creating a [DestinationRule](https://istio.io/docs/reference/config/networking/destination-rule/).  

- To enable sidecar injection for Pods of existing Services, restart them after upgrading to Kyma 1.0 or higher.
- To create a DestinationRule for a Service, follow the official [Istio documentation](https://istio.io/docs/reference/config/networking/destination-rule/).

>**TIP:** You can use the [`istioctl authn tls-check`](https://istio.io/docs/reference/commands/istioctl/#istioctl-authn-tls-check) command to find out where communication fails.
