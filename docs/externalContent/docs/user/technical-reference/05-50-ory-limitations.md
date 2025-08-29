# Default Configuration and Limitations

## APIRule Controller Limitations

APIRule Controller relies on Istio and Ory custom resources to provide routing capabilities. In terms of persistence, the controller depends only on APIRules stored in the Kubernetes cluster.
In terms of the resource configuration, the following requirements are set on APIGateway Controller:
- CPU Requests: 10m
- CPU Limits: 100m
- Memory Requests: 64Mi
- Memory Limits: 128Mi

You can create an unlimited number of APIRule custom resources.

## Ory Resources' Configuration

The configuration of Ory resources depends on the cluster capabilities. If your cluster has fewer than 5 total virtual CPU cores or its total memory capacity is less than 10 gigabytes, the default setup for resources is lighter. If your cluster exceeds both of these thresholds, the higher resource configuration is applied.

The default configuration for larger clusters includes the following settings for the Ory components' resources:

| Component          | CPU Requests | CPU Limits | Memory Requests | Memory Limits |
|--------------------|--------------|------------|-----------------|---------------|
| Oathkeeper         | 100m         | 10000m     | 64Mi            | 512Mi         |
| Oathkeeper Maester | 10m          | 400m       | 32Mi            | 1Gi           |

The default configuration for smaller clusters includes the following settings for the Ory components' resources:

| Component          | CPU Requests | CPU Limits | Memory Requests | Memory Limits |
|--------------------|--------------|------------|-----------------|---------------|
| Oathkeeper         | 10m          | 100m       | 64Mi            | 128Mi         |
| Oathkeeper Maester | 10m          | 100m       | 20Mi            | 50Mi          |


## Autoscaling Configuration

The default configuration in terms of autoscaling of Ory components is as follows:

| Component          | Min replicas | Max replicas |
|--------------------|--------------|--------------|
| Oathkeeper         | 3            | 10           |
| Oathkeeper Maester | 3            | 10           |

Oathkeeper Maester is a separate container running in the same Pod as Oathkeeper. Because of that, the autoscaling configuration of the Oathkeeper and Oathkeeper Master components is similar. The autoscaling configuration is based on CPU utilization, with HorizontalPodAutoscaler set up for 80% average CPU request utilization.