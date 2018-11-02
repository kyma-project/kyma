---
title: How to register a broker
type: Getting Started
---

This Getting Started guide shows how to register a new broker in the Service Catalog. The broker can be either a Namespace-scoped ServiceBroker or a cluster-wide ClusterServiceBroker. Follow this guide to register a cluster-wide [UPS Broker](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker) in the Service Catalog.

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)
* [helm](https://github.com/helm/helm#install)

## Steps

1. Fork the [`service-catalog`](https://github.com/kubernetes-incubator/service-catalog) repository and clone it to your local machine. To learn how, follow [this](https://github.com/kyma-project/community/blob/master/git-workflow.md#prepare-the-fork) instruction.

2.  Run this command to install the chart with the `ups-broker` name in the `ups-broker` Namespace:

  ```
helm install ./chart/ups-broker --name ups-broker --namespace ups-broker
```

3. Register your broker:
```
kubectl create -f contrib/examples/walkthrough/ups-broker.yaml
```
After you successfully register your ServiceBroker or ClusterServiceBroker, the Service Catalog periodically fetches services from this broker and creates ServiceClasses or ClusterServiceClasses from them.

4. Check the status of the broker:
```
kubectl get clusterservicebrokers ups-broker -o yaml
```
>**NOTE:** In case of ServiceBrokers, run:
>```
>kubectl get servicebrokers {name} -n {namespace} -o yaml
>```

  The output looks as follows:
```console
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  creationTimestamp: 2018-10-09T08:25:25Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  name: ups-broker
  resourceVersion: "10"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/ups-broker
  uid: deefbd1e-cb9c-11e8-8372-fade7e9a18e5
spec:
  relistBehavior: Duration
  relistRequests: 0
  url: http://ups-broker-ups-broker.ups-broker.svc.cluster.local
status:
  conditions:
  - lastTransitionTime: 2018-10-09T08:25:25Z
    message: Successfully fetched catalog entries from broker.
    reason: FetchedCatalog
    status: "True"
    type: Ready
  lastCatalogRetrievalTime: 2018-10-09T08:25:25Z
  reconciledGeneration: 1
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
