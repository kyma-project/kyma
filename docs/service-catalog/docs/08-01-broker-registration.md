---
title: Register a broker in the Service Catalog
type: Tutorials
---

This Getting Started guide shows how to register a broker in the Service Catalog. The broker can be either a Namespace-scoped ServiceBroker or a cluster-wide ClusterServiceBroker. Follow this guide to register a cluster-wide or Namespace-scoped version of the  [UPS Broker](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker).

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)
* [helm](https://github.com/helm/helm#install)

## Steps

1. Clone the [`service-catalog`](https://github.com/kubernetes-incubator/service-catalog) repository:
    ```
    git clone https://github.com/kubernetes-incubator/service-catalog.git
    ```

2.  Run this command to install the chart with the `ups-broker` name in the `stage` Namespace:
      ```
     helm install service-catalog/charts/ups-broker --name ups-broker --namespace stage
     ```

3. Register a broker:
  * Run this command to register a ClusterServiceBroker:
     ```
    kubectl create -f service-catalog/contrib/examples/walkthrough/ups-clusterservicebroker.yaml
    ```
  * To register the UPS Broker as a ServiceBroker in the `stage` Namespace, run:
    ```
    kubectl create -f service-catalog/contrib/examples/walkthrough/ups-servicebroker.yaml -n stage
    ```     
    After you successfully register your ServiceBroker or ClusterServiceBroker, the Service Catalog periodically fetches services from this broker and creates ServiceClasses or ClusterServiceClasses from them.

4. Check the status of your broker:
  * To check the status of your ClusterServiceBroker, run:
     ```
    kubectl get clusterservicebrokers ups-broker -o jsonpath="{.status.conditions}"
    ```
  * To check the status of the ServiceBroker, run:
    ```
    kubectl get servicebrokers ups-broker -n stage -o jsonpath="{.status.conditions}"
    ```

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

5. View Service Classes that this broker provides:
  * To check the ClusterServiceClasses, run:
      ```
     kubectl get clusterserviceclasses
      ```
  * To check the ServiceClasses, run:
      ```
      kubectl get serviceclasses -n stage
      ```

      These are the UPS Broker Service Classes:
      ```
      NAME                                   EXTERNAL NAME
      4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468   user-provided-service
      5f6e6cf6-ffdd-425f-a2c7-3c9258ad2468   user-provided-service-single-plan
      8a6229d4-239e-4790-ba1f-8367004d0473   user-provided-service-with-schemas
      ```
