---
title: How to register a broker
type: Getting Started
---

This Getting Started guide shows how to register a new broker in the Service Catalog. The broker can be either a Namespace-scoped ServiceBroker or a cluster-wide ClusterServiceBroker. Follow the instructions based on the cluster-wide [UPS Broker](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker) to complete the guide.

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)
* broker's `yaml` files that specify:
  * application which implements the [Open Service Broker API](https://www.openservicebrokerapi.org/)
  * Kubernetes service which enables the connection between a broker and an application
  * broker registration file in which the kind of a broker is specified

> **NOTE:** In case of the sample UPS Broker, find the application and service files [here](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker/templates). Use [this](https://github.com/kubernetes-incubator/service-catalog/blob/master/contrib/examples/walkthrough/ups-broker.yaml) registration file.

## Steps

Run these commands to register the UPS broker using kubectl:

```
kubectl apply -f [broker-deployment.yaml](https://raw.githubusercontent.com/kubernetes-incubator/service-catalog/master/charts/ups-broker/templates/broker-deployment.yaml) -n ups-broker
kubectl apply -f [broker-service.yaml](https://raw.githubusercontent.com/kubernetes-incubator/service-catalog/master/charts/ups-broker/templates/broker-service.yaml) -n ups-broker
kubectl apply -f [ups-broker.yaml](https://raw.githubusercontent.com/kubernetes-incubator/service-catalog/master/contrib/examples/walkthrough/ups-broker.yaml)
```
>**NOTE:** In case of ServiceBrokers, specify a Namespace by adding ``-n {namespace}`` to the last command.

### Check the status

To check if your broker's registration is successful, run:

```
kubectl get clusterservicebroker -n ups-broker ups-broker -o jsonpath="{.status.conditions}"
```

If the broker is registered successfully, the output looks as follows:

```
{
    "lastTransitionTime": "2018-10-26T12:03:32Z",
    "message": "Successfully fetched catalog entries from broker.",
    "reason": "FetchedCatalog",
    "status": "True",
    "type": "Ready"
}
```

After you successfully register your ServiceBroker or ClusterServiceBroker, the Service Catalog periodically fetches services from this broker and creates ServiceClasses or ClusterServiceClasses from them.

To check the created ClusterServiceClasses, run this command:
```
kubectl get clusterserviceclasses
```

>**NOTE:** In case of ServiceClasses, the command looks as follows:
>
>```
>kubectl get serviceclasses -n {namespace}
>```
