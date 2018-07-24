---
title: Ingress-Gateway controller
type: Architecture
---

The Ingress-Gateway controller exposes the Kyma gateways to the outside world by the public IP address/DNS name.
The DNS name of the Ingress is `gateway.[cluster-dns]`. For example: `gateway.servicemanager.cluster.kyma.cx`.

A particular Remote Environment is exposed as a path. For example, to reach the Gateway for the Remote Environment named `ec-default`, use the following URL: `gateway.servicemanager.cluster.kyma.cx/ec-default`

This is an example of how to get all ServiceClasses:

```console
http GET https://gateway.servicemanager.cluster.kyma.cx/ec-default/v1/metadata/services --cert=ec-default.pem
```
