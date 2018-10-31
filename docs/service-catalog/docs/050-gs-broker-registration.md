---
title: How to register a broker
type: Getting Started
---

This Getting Started guide shows how to register a new broker in the Service Catalog. The broker can be either a Namespace-scoped ServiceBroker or a cluster-wide ClusterServiceBroker. Follow the instructions based on the [UPS Broker example](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker) to complete the guide.

## Prerequisites

* [Service Catalog](https://github.com/kubernetes-incubator/service-catalog/releases) running in version `0.1.28` or higher
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl) or [helm](https://github.com/helm/helm#install) installed
* The `yaml` files for the broker that specify:
  * an application which implements the [Open Service Broker API](https://www.openservicebrokerapi.org/)
  * a Kubernetes service which enables the connection between a broker and an application
  * broker registration file in which the kind of a broker is specified

> **NOTE:** In case of the sample UPS Broker, find the application and service files [here](https://github.com/kubernetes-incubator/service-catalog/tree/master/charts/ups-broker/templates). Use [this](https://github.com/kubernetes-incubator/service-catalog/blob/master/contrib/examples/walkthrough/ups-broker.yaml) registration file.

## Steps

You can register your broker in the Service Catalog either by installing a Helm chart or by using kubectl commands. You can also register it directly using the Kyma Console.

### Register using Helm chart

To register a broker in the Service Catalog using a Helm chart, go to the chart's directory and run this command:

```
helm install {chart directory} --name {broker name} --namespace {namespace}
```
For example, run this command to install the chart with the `ups-broker` name in the `ups-broker` Namespace:

```
helm install charts/ups-broker --name ups-broker --namespace ups-broker
```

### Register using kubectl

Run these commands to register a broker using kubectl:
```
kubectl apply -f {application filename} -n {namespace}
kubectl apply -f {service filename} -n {namespace}
kubectl apply -f {broker registration filename} -n {namespace}
```
In case of the UPS Broker, the commands look as follows:
```
kubectl apply -f broker-deployment.yaml -n ups-broker
kubectl apply -f broker-service.yaml -n ups-broker
kubectl apply -f ups-broker.yaml -n ups-broker
```

### Register using the Console

1. Go to the Kyma Console and choose the Environment.
2. Click the **Deploy new resource to the environment** button.
3. Select broker's `yaml` files and click **Upload**.

>**NOTE:** This method applies only to ServiceBrokers. You cannot register ClusterServiceBrokers using the Kyma Console.

### Check the status

To check if your broker's registration is successful, run:

```
kubectl get servicebroker -n {namespace} {broker name} -o jsonpath="{.status.conditions}"
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
