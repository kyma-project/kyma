---
title: Nginx Ingress Controller
type: Details
---

The Nginx Ingress Controller controller exposes the Application Connector to the outside world by the public IP address/DNS name.
The DNS name of the Ingress is `gateway.[cluster-dns]`. For example: `gateway.servicemanager.cluster.kyma.cx`.

A particular Remote Environment is exposed as a path. For example, to reach the Gateway for the Remote Environment named `ec-default`, use the following URL: `gateway.servicemanager.cluster.kyma.cx/ec-default`

The Nginx Ingress Controller is protecting endpoint with a certificate validation. Each calls must be done with a proper client certificate which is aquired for a Remote Environment.
You can check more details about client certificate in the following document: [Connector Service](TODO)

