---
title: How to register a broker
type: Getting Started
---

This Getting Started guide shows how to register a new broker in the Service Catalog. The broker can be either a Namespace-scoped ServiceBroker or a cluster-wide ClusterServiceBroker. Follow this guide to register a cluster-wide [UPS Broker](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker) in the Service Catalog.

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)
* [helm](https://github.com/helm/helm#install)

## Steps

1. Clone the [`service-catalog`](https://github.com/kubernetes-incubator/service-catalog) repository to your local machine:
```
git clone https://github.com/kubernetes-incubator/service-catalog.git
```

2.  Run this command to install the chart with the `ups-broker` name in the `ups-broker` Namespace:

  ```
helm install service-catalog/charts/ups-broker --name ups-broker --namespace ups-broker
```

3. Register your broker:
```
kubectl create -f service-catalog/contrib/examples/walkthrough/ups-broker.yaml
```
After you successfully register your ServiceBroker or ClusterServiceBroker, the Service Catalog periodically fetches services from this broker and creates ServiceClasses or ClusterServiceClasses from them.

4. Check the status of the broker:
```
kubectl get clusterservicebrokers ups-broker -o jsonpath="{.status.conditions}"
```
>**NOTE:** In case of ServiceBrokers, run:
>```
>kubectl get servicebrokers {name} -n {namespace} -o jsonpath="{.status.conditions}"
>```

  The output looks as follows:
```
{
    "lastTransitionTime": "2018-10-26T12:03:32Z",
    "message": "Successfully fetched catalog entries from broker.",
    "reason": "FetchedCatalog",
    "status": "True",
    "type": "Ready"
}
  ```

5. View ClusterServiceClasses that this broker provides:
  ```
kubectl get clusterserviceclasses
```
>**NOTE:** In case of ServiceBrokers, the command looks as follows:
>
>```
>kubectl get serviceclasses -n {namespace}
>```

  These are the UPS Broker ClusterServiceClasses:
```
NAME                                   EXTERNAL NAME
4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468   user-provided-service
5f6e6cf6-ffdd-425f-a2c7-3c9258ad2468   user-provided-service-single-plan
8a6229d4-239e-4790-ba1f-8367004d0473   user-provided-service-with-schemas
```
