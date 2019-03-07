---
title: CLI reference
---

Management of the Service Catalog is based on Kubernetes resources and the custom resources specifically defined for Kyma. Manage all of these resources through [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/).

## Details

This section describes the resource names to use in the kubectl command line, the command syntax, and examples of use.

### Resource types

Service Catalog operations use the following resources:

| Singular name      | Plural name         |
| ------------------ |---------------------|
|clusterservicebroker|clusterservicebrokers|
|clusterserviceclass |clusterserviceclasses|
|clusterserviceplan  |clusterserviceplans  |
|secret              |secrets              |
|servicebinding      |servicebindings      |
|servicebindingusage |servicebindingusages |
|servicebroker       |servicebrokers       |
|serviceclass        |serviceclasses       |
|serviceinstance     |serviceinstances     |
|serviceplan         |serviceplans         |


### Syntax

Follow the `kubectl` syntax, `kubectl {command} {type} {name} {flags}`, where:

* {command} is any command, such as `describe`.
* {type} is a resource type, such as `clusterserviceclass`.
* {name} is the name of a given resource type. Use {name} to make the command return the details of a given resource.
* {flags} specifies the scope of the information. For example, use flags to define the Namespace from which to get the information.

### Examples
The following examples show how to create a ServiceInstance, how to get a list of ClusterServiceClasses and a list of ClusterServiceClasses with human-readable names, a list of ClusterServicePlans, and a list of all ServiceInstances.

* Create a ServiceInstance using the example of the Redis ServiceInstance for the 0.1.38 version of the Service Catalog:

```
cat <<EOF | kubectl create -f -
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: my-instance
  namespace: stage

spec:
  clusterServiceClassExternalName: redis
  clusterServicePlanExternalName: micro
  parameters:
     "imagePullPolicy": "Always"
EOF
```

* Get the list of all ClusterServiceClasses:
```
kubectl get clusterserviceclasses
```
* Get the list of all ClusterServiceClasses and their human-readable names:
```
kubectl get clusterserviceclasses -o=custom-columns=NAME:.metadata.name,EXTERNAL\ NAME:.spec.externalName
```

* Get the list of all ClusterServicePlans and associated ClusterServiceClasses:
```
kubectl get clusterserviceplans -o=custom-columns=NAME:.metadata.name,EXTERNAL\ NAME:.spec.externalName,EXTERNAL\ SERVICE\ CLASS:.spec.clusterServiceClassRef
```
* Get the list of all ServiceInstances from all Namespaces:
```
kubectl get serviceinstances --all-namespaces
```
