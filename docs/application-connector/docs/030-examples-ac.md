---
title: Examples
type: Examples
---

This is a tutorial on how to get a new Application Connector running on Minikube.

To integrate a new Remote Environment marked as `Production`, you can use the following values:
* **sourceEnvironment** = `production`
* **sourceType** = `marketing`
* **sourceNamespace** = `organization.com`

Start with:

``` bash
helm install --name hmc-prod --set deployment.args.sourceType=marketing --set deployment.args.sourceEnvironment=production --set global.isLocalEnv=true --set service.externalapi.nodePort=32002 --namespace kyma-integration ./remote-environments
```

Your output looks like this:
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
Running `helm status hmc-prod` shows a similar output with the most recent status of the release.

If you run `helm list`, you can see your release among the others:
``` bash
cluster-essentials        	1       	Wed Apr 18 07:50:01 2018	DEPLOYED	kyma-cluster-essentials-0.0.1 kyma-system
ec-default                	1       	Wed Apr 18 07:57:50 2018	DEPLOYED	gateway-0.0.1              	  kyma-integration
hmc-default               	1       	Wed Apr 18 07:57:36 2018	DEPLOYED	gateway-0.0.1              	  kyma-integration
istio                     	1       	Wed Apr 18 07:50:04 2018	DEPLOYED	istio-0.5.0                	  istio-system
prometheus-operator       	1       	Wed Apr 18 07:51:50 2018	DEPLOYED	prometheus-operator-0.17.0 	  kyma-system
hmc-prod                  	1       	Fri Apr 20 11:25:44 2018	DEPLOYED	gateway-0.0.1              	  kyma-integration
core                      	2       	Wed Apr 18 07:56:56 2018	DEPLOYED	core-0.0.1                 	  kyma-system
```

Use kubectl commands to see the Kubernetes resources associated with your release:

```
kubectl get pods -n kyma-integration
```

``` bash
NAME                                                  READY     STATUS      RESTARTS   AGE
ec-default-gateway-5b77fdf7b5-rx64m                   2/2       Running     3          2d
hmc-default-gateway-f88b58978-75dkb                   2/2       Running     3          2d
hmc-prod-gateway-67469769c8-6lgjl                     1/1       Running     0          1m
```

```
kubectl get services -n kyma-integration
```

``` bash
NAME                                              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
ec-default-gateway-echo                           ClusterIP   10.96.212.205    <none>        8080/TCP         2d
ec-default-gateway-external-api                   NodePort    10.101.245.196   <none>        8081:32000/TCP   2d
hmc-default-gateway-echo                          ClusterIP   10.101.68.223    <none>        8080/TCP         2d
hmc-default-gateway-external-api                  NodePort    10.96.215.1      <none>        8081:32001/TCP   2d
hmc-prod-gateway-echo                             ClusterIP   10.100.94.12     <none>        8080/TCP         1m
hmc-prod-gateway-external-api                     NodePort    10.108.126.243   <none>        8081:32002/TCP   1m
```

If you are done with the release, you can delete it by following this example:
```
helm delete hmc-prod --purge
```

```bash
release "hmc-prod" deleted
```
