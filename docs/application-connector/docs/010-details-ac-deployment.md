---
title: Deploy a new Application Connector
type: Details
---

By default, Kyma comes with two Application Connectors preconfigured. Those Application Connectors are installed in the `kyma-integration` Namespace.


### Install a Application Connector locally

For installation on Minikube, provide the NodePort as shown in this example:

``` bash
helm install --name {remote-environment-name} --set deployment.args.sourceType=commerce --set global.isLocalEnv=true --set service.externalapi.nodePort=32001 --namespace kyma-integration ./resources/remote-environments
```

You can override the following parameters:

- **sourceEnvironment** - the Event source environment name.
- **sourceType** - the Event source type.
- **sourceNamespace** - the organization that publishes the Event.


### Install an Application Connector on the cluster

To add a new Application Connector to the cluster, download [remote-environments.zip](assets/remote-environments.zip) package, unpack it, and place the content in the project's directory.

To install a Remote Environment, use:
``` bash
helm install --name {remote-environment-name} --set deployment.args.sourceType=commerce --set global.isLocalEnv=false --set global.domainName={domain-name} --namespace kyma-integration ./remote-environments
```

- global.domainName override is required and cannot be omitted, example values may look like:
```
wormhole.cluster.kyma.cx
nightly.cluster.kyma.cx
```

You can override the following parameters:

- **sourceEnvironment** is the Event source environment name.
- **sourceType** is the Event source type.
- **sourceNamespace** is the organization that publishes the Event.

### Working with Helm

Helm provides the following commands:
- `helm list` - lists existing Helm releases
- `helm test [release-name]` - tests a release
- `helm get [release-name]` - shows the contents of `.yaml` files that make up the release
- `helm status [release-name]` - shows the status of a named release
- `helm delete [release-name]` - deletes a release from Kubernetes

The full list of the Helm commands is available in the [Helm documentation](https://docs.helm.sh/helm/).
You can also use the `helm --help` command.

### Use kubectl

To check if everything runs correctly, use kubectl:
`kubectl get pods -n kyma-integration`  
`kubectl get services -n kyma-integration`  

### Examples

Follow the **Running a new Application Connector on Minikube** tutorial to learn how to get a new Application Connector running on Minikube.
