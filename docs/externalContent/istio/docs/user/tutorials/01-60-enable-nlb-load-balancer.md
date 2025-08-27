# Migrate From Using Elastic Load Balancer (ELB) to Network Load Balancer (NLB) for the Istio Module Running on AWS

> [!WARNING]
>
> Switching the load balancer type may cause brief downtime for the Istio Ingress Gateway.
> Make sure to plan the migration process accordingly,
> in a maintenance window that minimizes the impact on the application availability.
> The migration process from ELB to NLB is irreversible.

## Introduction

Until version 1.16, the Istio module used the Elastic Load Balancer (ELB) as the load balancer type for the Istio Ingress Gateway.
Starting from version 1.16.0, the Network Load Balancer (NLB) is used as the new default.
This change was made to improve the feature compatibility with the AWS environment,
as well as make the Istio module's installation more uniform across different cloud providers.

To facilitate safe migration from ELB to NLB, 
version 1.16 of the module creates the `elb-deprecated` ConfigMap in the `istio-system` namespace.
This ConfigMap safeguards against downtime happening during the upgrade process to 1.16,
making sure that the ELB is still used as the load balancer type for the Istio Ingress Gateway as long as the ConfigMap is present.

## Migration

To migrate from using ELB to NLB for the Istio module running on AWS, follow these steps:
1. Make sure that you are using the 1.16 or later version of the module.
2. Remove the `elb-deprecated` ConfigMap from the `istio-system` namespace.
3. The module automatically switches to using the NLB as the load balancer type for the Istio Ingress Gateway.
