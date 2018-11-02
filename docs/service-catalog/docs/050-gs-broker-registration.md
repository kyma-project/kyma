---
title: How to register a broker
type: Getting Started
---

This Getting Started guide shows how to register a new broker in the Service Catalog. The broker can be either a Namespace-scoped ServiceBroker or a cluster-wide ClusterServiceBroker. Follow this guide to register a cluster-wide [UPS Broker](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker) in the Service Catalog.

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)
* [helm](https://github.com/helm/helm#install)

## Steps

1. Fork the [`service-catalog`](https://github.com/kubernetes-incubator/service-catalog) repository and clone it to your local machine:
```
git clone https://github.com/{your-username}/{your-fork-name}.git
```

2.  Run this command to install the chart with the `ups-broker` name in the `ups-broker` Namespace:

  ```
helm install charts/ups-broker --name ups-broker --namespace ups-broker
```

3. Check the status

  ```
kubectl get clusterserviceclasses
```

  >**NOTE:** In case of ServiceBrokers, the command looks as follows:
>
>```
>kubectl get serviceclasses -n {namespace}
>```

After you successfully register your ServiceBroker or ClusterServiceBroker, the Service Catalog periodically fetches services from this broker and creates ServiceClasses or ClusterServiceClasses from them.
