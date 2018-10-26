---
title: How to register a broker
type: Getting Started
---

This Getting Started guide shows how to register a new broker to the Service Catalog. Follow the example of the [UPS Broker](https://github.com/kyma-project/kyma/tree/master/tests/ui-api-layer-acceptance-tests/domain/servicecatalog/testdata/charts/ups-broker) to complete the guide.

## Prerequisites

To register a new broker to the Service Catalog, you must have:
* Service Catalog running in version `0.1.28` or higher
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl) or [helm](https://github.com/helm/helm#install) installed
* broker's `yaml` files that specify:
  * application which implements the [Open Service Broker API](https://www.openservicebrokerapi.org/)
  * broker registration
  * Kubernetes service which enables the connection between a broker and an application

## Steps

You can register your broker to the Service Catalog either by installing a Helm chart or by using kubectl commands. You can also register it directly using the Kyma Console.

### Register using Helm chart

To register a broker to the Service Catalog using a Helm chart, go to the chart's directory and run this command:

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
kubectl apply -f broker-deployment.yaml -n qa
kubectl apply -f broker-service.yaml -n qa
kubectl apply -f broker-register.yaml -n qa
```
>**NOTE:** In case of a ClusterServiceBroker, do not specify the Namespace and skip the `-n {namespace}` part of the command.

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

After you successfully register your ServiceBroker or ClusterServiceBroker, the Service Catalog periodically fetches the services from this broker and creates ServiceClasses or ClusterServiceClasses from them.
