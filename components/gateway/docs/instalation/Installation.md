# Adding a new Application Connector to Kyma

## Overview

The Application Connector connects an external solution to Kyma.

## Introduction

The Application Connector consists of:
- the [Remote Environment](https://github.com/kyma-project/kyma/blob/master/docs/remote-environment.md)
- the Gateway - A service responsible for registering available services (APIs, Events) and proxying calls to the registered solution.
- Ingress-Nginx - A controller that exposes multiple Application Connectors to the external world.

By default, Kyma comes with two default Application Connectors preconfigured. A user can add more Application Connectors using Helm package manager.

To add an Application Connector, download the `remote-environments.zip` package. Unpack it and place it in the project directory.


## Installation

Use this command to install the Remote Environment:
``` bash
helm install --name remote-environment-name --set deployment.args.sourceType=commerce --set global.domainName=domain.cluster.com --set global.isLocalEnv=false --namespace kyma-integration ./remote-environments
```

To install locally on Minikube, provide the NodePort as shown in this example:
``` bash
helm install --name remote-environment-name --set deployment.args.sourceType=commerce --set global.domainName=domain.cluster.com --set global.isLocalEnv=true --set service.externalapi.nodePort=32001 --namespace kyma-integration ./remote-environments
```

The user can override the following parameters:

- **sourceEnvironment** - The Event source environment name.
- **sourceType** - The Event source type.
- **sourceNamespace** - The organization that publishes the Event.

## Working with helm

Helm provides you with several useful commands:
- `helm list` - list existing helm releases
- `helm test [release-name]` - test a release
- `helm get [release-name]` - see the contents of `.yaml` files that make up the release
- `helm status [release-name]` - show the status of the named release
- `helm delete [release-name]` - delete the release from Kubernetes

To review a complete list of Helm commands, see [Helm documentation](https://docs.helm.sh/helm/) or use `helm --help`

 ## Check with kubectl

Make sure everything runs with kubectl:
`kubectl get pods -n kyma-integration`
`kubectl get services -n kyma-integration`

## Access the Application Connector

The Ingress-Nginx controller exposes Kyma Gateways to the outside world using a public IP address/DNS name. The DNS name of Ingress is `gateway.[cluster-dns]`. For example, `gateway.servicemanager.cluster.kyma.cx`.

Expose a particular Gateway service as the path of the Remote Environment. For example, if you want to reach the Gateway of the Remote Environment named `ec-dafault`, you need to use following URL: `gateway.servicemanager.cluster.kyma.cx/ec-default`. The communication requires a valid client certificate. Check the security documentation for further details.

The following example shows how to get all ServiceClasses:

``` console
http GET https://gateway.servicemanager.cluster.kyma.cx/ec-default/v1/metadata/services --cert=ec-default.pem
```

## Example

This example shows how to add a new Application Connector to the Minikube.

To integrate a new instance of `Marketing` marked as a `Production` environment, the example uses the following values:

- **sourceEnvironment** = production
- **sourceType** = marketing
- **sourceNamespace** = organization.com

Start with:

``` bash
helm install --name hmc-prod --set deployment.args.sourceType=marketing --set deployment.args.sourceEnvironment=production --set global.isLocalEnv=true --set service.externalapi.nodePort=32002 --namespace kyma-integration ./remote-environments
```

The following output displays:
``` bash
NAME:   hmc-prod                  
LAST DEPLOYED: Fri Apr 20 11:25:44 2018
NAMESPACE: kyma-integration
STATUS: DEPLOYED

RESOURCES:
==> v1/Role
NAME                   AGE
hmc-prod-gateway-role  0s

==> v1/RoleBinding
NAME                          AGE
hmc-prod-gateway-rolebinding  0s

==> v1/Service
NAME                           TYPE       CLUSTER-IP      EXTERNAL-IP  PORT(S)         AGE
hmc-prod-gateway-external-api  NodePort   10.108.126.243  <none>       8081:32002/TCP  0s
hmc-prod-gateway-echo          ClusterIP  10.100.94.12    <none>       8080/TCP        0s

==> v1beta1/Deployment
NAME              DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
hmc-prod-gateway  1        1        1           0          0s

==> v1alpha1/RemoteEnvironment
NAME      AGE
hmc-prod  0s

==> v1/Pod(related)
NAME                               READY  STATUS             RESTARTS  AGE
hmc-prod-gateway-67469769c8-6lgjl  0/1    ContainerCreating  0         0s


NOTES:
------------------------------------------------------------------------------------------------------------------------

Thank you for installing Gateway helm chart for Kubernetes version 0.0.1.

To learn more about the release, see:

  $ helm status hmc-prod                  
  $ helm get hmc-prod                  

------------------------------------------------------------------------------------------------------------------------

```
Running `helm status hmc-prod` shows similar output with the most recent status of the release.

Run the `helm list` command. See your release among the others:
``` bash
cluster-essentials        	1       	Wed Apr 18 07:50:01 2018	DEPLOYED	kyma-cluster-essentials-0.0.1	kyma-system
ec-default                	1       	Wed Apr 18 07:57:50 2018	DEPLOYED	gateway-0.0.1              	    kyma-integration
hmc-default               	1       	Wed Apr 18 07:57:36 2018	DEPLOYED	gateway-0.0.1              	    kyma-integration
istio                     	1       	Wed Apr 18 07:50:04 2018	DEPLOYED	istio-0.5.0                	    istio-system
prometheus-operator       	1       	Wed Apr 18 07:51:50 2018	DEPLOYED	prometheus-operator-0.17.0 	    kyma-system
hmc-prod                  	1       	Fri Apr 20 11:25:44 2018	DEPLOYED	gateway-0.0.1              	    kyma-integration
sf-core                   	2       	Wed Apr 18 07:56:56 2018	DEPLOYED	kyma-core-0.0.1              	kyma-system
```

Use `kubectl` commands to see Kubernetes resources associated with the release.  

`kubectl get pods -n kyma-integration`
``` bash
NAME                                                  READY     STATUS      RESTARTS   AGE
ec-default-gateway-5b77fdf7b5-rx64m                   2/2       Running     3          2d
hmc-default-gateway-f88b58978-75dkb                   2/2       Running     3          2d
hmc-prod-gateway-67469769c8-6lgjl                     1/1       Running     0          1m
```

`kubectl get services -n kyma-integration`
``` bash
NAME                                              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
ec-default-gateway-echo                           ClusterIP   10.96.212.205    <none>        8080/TCP         2d
ec-default-gateway-external-api                   NodePort    10.101.245.196   <none>        8081:32000/TCP   2d
hmc-default-gateway-echo                          ClusterIP   10.101.68.223    <none>        8080/TCP         2d
hmc-default-gateway-external-api                  NodePort    10.96.215.1      <none>        8081:32001/TCP   2d
hmc-prod-gateway-echo                             ClusterIP   10.100.94.12     <none>        8080/TCP         1m
hmc-prod-gateway-external-api                     NodePort    10.108.126.243   <none>        8081:32002/TCP   1m
```

When you are done, delete the release with the following command:
`helm delete hmc-prod --purge`

```bash
release "hmc-prod" deleted
```
