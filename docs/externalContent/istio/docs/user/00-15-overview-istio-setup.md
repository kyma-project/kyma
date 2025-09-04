# Default Istio Configuration

Within the Istio module, Istio Operator provides baseline values for the Istio installation, which you can override in the Istio custom resource (CR).

See the major differences in the configuration of Istio Operator compared to upstream Istio:

- Istiod (Pilot) and Ingress Gateway components are enabled by default.
- Automatic Istio sidecar proxy injection is disabled by default.
- To enhance security and performance, both [Istio control plane and data plane](https://istio.io/latest/docs/ops/deployment/architecture/) use the distroless version of Istio images. Those images are not Debian-based and are slimmed down to reduce any potential attack surface. To learn more, see [Harden Docker Container Images](https://istio.io/latest/docs/ops/configuration/security/harden-docker-images/).
- Resource requests and limits for Istio sidecars proxies are modified to best suit the needs of the evaluation and production profiles.
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled in the STRICT mode for workloads in the Istio service mesh.
- Egress traffic is not controlled. All applications deployed in the Kyma cluster can access outside resources without limitations.
- The CNI component, used for the installation of an Istio sidecar, is provided as a DaemonSet. This means that one replica is present on every node of the target cluster.
- The self-signed CA certificate’s bit length is set to `4096` instead of the default `2048`.

## Configuration Based on the Cluster Size

The configuration of Istio resources depends on the cluster capabilities. If your cluster has less than 5 total virtual CPU cores or its total memory capacity is less than 10 gigabytes, the default setup for resources and autoscaling is lighter. If your cluster exceeds both of these thresholds, Istio is installed with the higher resource configuration.

### Default Resource Configuration for Smaller Clusters

| Component       | CPU Requests | CPU Limits | Memory Requests | Memory Limits |
|-----------------|--------------|------------|-----------------|---------------|
| Proxy           | 10 m         | 250 m      | 32 Mi           | 254 Mi        |
| Ingress Gateway | 10 m         | 1000 m     | 32 Mi           | 1024 Mi       |
| Egress Gateway  | 10 m         | 1000 m     | 32 Mi           | 1024 Mi       |
| Pilot           | 50 m         | 1000 m     | 128 Mi          | 1024 Mi       |
| CNI             | 10 m         | 250 m      | 128 Mi          | 384 Mi        |

### Default Resource Configuration for Larger Clusters

| Component       | CPU Requests | CPU Limits | Memory Requests | Memory Limits |
|-----------------|--------------|------------|-----------------|---------------|
| Proxy           | 10 m         | 1000 m     | 192 Mi          | 1024 Mi       |
| Ingress Gateway | 100 m        | 2000 m     | 128 Mi          | 1024 Mi       |
| Egress Gateway  | 100 m        | 2000 m     | 128 Mi          | 1024 Mi       |
| Pilot           | 100 m        | 4000 m     | 512 Mi          | 2 Gi          |
| CNI             | 100 m        | 500 m      | 512 Mi          | 1024 Mi       |

### Default Autoscaling Configuration for Smaller Clusters

| Component       | minReplicas | maxReplicas |
|-----------------|-------------|-------------|
| Ingress Gateway | 1           | 1           |
| Egress Gateway  | 1           | 1           |
| Pilot           | 1           | 1           |

### Default Autoscaling Configuration for Larger Clusters

| Component       | minReplicas | maxReplicas |
|-----------------|-------------|-------------|
| Ingress Gateway | 3           | 10          |
| Egress Gateway  | 3           | 10          |
| Pilot           | 2           | 5           |
