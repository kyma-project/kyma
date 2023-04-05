---
title: Limitations of Istio Service Mesh
---

## Resource configuration

By default, Istio resources are configured in the following matter:

| Component       |          | CPU   | Memory |
|-----------------|----------|-------|--------|
| Proxy           | Limits   | 250m  | 256Mi  |
| Proxy           | Requests | 10m   | 64Mi   |
| Ingress Gateway | Limits   | 2000m | 1024Mi |
| Ingress Gateway | Requests | 100m  | 128Mi  |
| Pilot           | Limits   | 500m  | 1024Mi |
| Pilot           | Requests | 100m  | 512Mi  |
| CNI             | Limits   | 500m  | 1024Mi |
| CNI             | Requests | 100m  | 512Mi  |

## Autoscaling configuration

The autoscaling configuration for Istio components is as follows:

| Component       | Min replicas | Max replicas |
|-----------------|--------------|--------------|
| Pilot           | 1            | 5            |
| Ingress Gateway | 1            | 5            |

`CNI` component is provided as a `DaemonSet` meaning that one replica will be present on every node of target cluster. `Proxy` doesn't have any configuration in terms of autoscaling as it is deployed by injecting a `Pod` with [sidecar injection enabled](../../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).
