---
title: Limitations of Istio Service Mesh
---

## Resource configuration

By default, Istio resources are configured in the following matter:

| Component       |          | CPU   | Memory |
|-----------------|----------|-------|--------|
| Proxy           | Limits   | 1000m | 1024Mi |
| Proxy           | Requests | 10m   | 192Mi  |
| Ingress Gateway | Limits   | 2000m | 1024Mi |
| Ingress Gateway | Requests | 100m  | 128Mi  |
| Pilot           | Limits   | 4000m | 2Gi    |
| Pilot           | Requests | 100m  | 512Mi  |
| CNI             | Limits   | 500m  | 1024Mi |
| CNI             | Requests | 100m  | 512Mi  |

## Autoscaling configuration

The autoscaling configuration of the Istio components is as follows:

| Component       | Min replicas | Max replicas |
|-----------------|--------------|--------------|
| Pilot           | 2            | 5            |
| Ingress Gateway | 3            | 10           |

The CNI component is provided as a DaemonSet, meaning that one replica is present on every node of the target cluster. Istio sidecar proxy isn't configured in terms of autoscaling as it is injected into a Pod with the [sidecar injection enabled](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).
