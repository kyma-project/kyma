---
title: Deploy a new Remote Environment
type: Details
---

By default, Kyma comes with two Remote Environments preinstalled. Those Remote Environments are installed in the `kyma-integration` Namespace.

>**NOTE:** A single instance of Application Connector allows you to connect one Remote Environment to Kyma. A Remote Environment is a representation of a connected external solution.

## Install a Remote Environment on a local Kyma deployment

To install a new Remote Environment on Minikube, provide the NodePort as shown in this example:

```
helm install --name {remote-environment-name} --set deployment.args.sourceType=commerce --set global.isLocalEnv=true --set service.externalapi.nodePort=32001 --namespace kyma-integration ./resources/remote-environments
```

You can override the following parameters:

- **sourceEnvironment** is the Event source environment name
- **sourceType** is the Event source type
- **sourceNamespace** is the organization that publishes the Event

Follow the **Set up a Remote Environment on Minikube** getting started guide to learn more about installing and setting up a Remote Environment on
a local Kyma installation.

## Install a Remote Environment on a cluster Kyma deployment

To add a new Remote Environment to the cluster, run this command:

```
helm install --name {remote-environment-name} --set deployment.args.sourceType=commerce --set global.isLocalEnv=false --set global.domainName={domain-name} --namespace kyma-integration ./resources/remote-environments
```

The **global.domainName** is mandatory. Example values can look like:
```
wormhole.cluster.kyma.cx
nightly.cluster.kyma.cx
```

You can override the following parameters:

- **sourceEnvironment** is the Event source environment name
- **sourceType** is the Event source type
- **sourceNamespace** is the organization that publishes the Event

## Working with Helm

Helm provides the following commands:
- `helm list` lists existing Helm releases
- `helm test [release-name]` tests a release
- `helm get [release-name]` shows the contents of `.yaml` files that make up the release
- `helm status [release-name]` shows the status of a named release
- `helm delete [release-name]` deletes a release from Kubernetes

The full list of the Helm commands is available in the [Helm documentation](https://docs.helm.sh/helm/).
You can also use the `helm --help` command.

## Use kubectl

To check if everything runs correctly, use the `kubectl get pods -n kyma-integration` or `kubectl get services -n kyma-integration` command.  
