# Testing scenarios

## Overview

This document shows possible scenarios for the Service Binding Controller.

Execute all commands from the examples in the `examples` directory.

## Usage

This section shows how to use bindings both on a Deployment and a Function.
The [Function](#use-bindings-on-a-function) scenario uses the prefixing functionality for the injected environment variable.  

### Use bindings on a Deployment

To use bindings on a Deployment, follow these steps:

1. Export the name of the Namespace.
```bash
export namespace="kyma-examples"
```
2. Create a Namespace.
```bash
kubectl create ns $namespace
```
3. Create a Redis instance.
```bash
kubectl create -f servicecatalog/redis-instance.yaml -n $namespace
```
4. Create Secrets for the Redis instance.
```bash
kubectl create -f servicecatalog/redis-instance-binding.yaml -n $namespace
```
5. Check if the Redis instance is already provisioned.
```bash
kubectl get serviceinstance/redis -n $namespace -o jsonpath='{ .status.conditions[0].reason }'
```
6. Create a Redis client.
```bash
kubectl create -f deploy/redis-client.yaml -n $namespace
```
7. Create a Binding Usage.
```bash
kubectl create -f deploy/service-binding-usage.yaml -n $namespace
```
8. Wait until the deployment Pod is ready.
```bash
kubectl get po -l app=redis-client -n $namespace -o jsonpath='{ .items[*].status.conditions[?(@.type=="Ready")].status }'
```
10. Export the name of the Pod.
```bash
export podName=$(kubectl get po -l app=redis-client -n $namespace -o jsonpath='{ .items[*].metadata.name }')
```
11. Execute the `check-redis` script on the Pod.
```bash
kubectl exec $podName -n $namespace /check-redis.sh
```
The information and statistics about the Redis server appear.

>**NOTE:** You can complete steps one to five through the UI.

To perform a clean-up, remove the Namespace:

```bash
kubectl delete ns $namespace
```

### Use bindings on a Function

To use bindings on a Function, follow these steps:

1. Export the name of the Namespace.
```bash
export namespace="kyma-examples"
```
2. Create a Namespace.
```bash
kubectl create ns $namespace
```
3. Create a Redis instance.
```bash
kubectl create -f servicecatalog/redis-instance.yaml -n $namespace
```
4. Create Secrets for the Redis instance.
```bash
kubectl create -f servicecatalog/redis-instance-binding.yaml -n $namespace
```
5. Check if the Redis instance is already provisioned.
```bash
kubectl get serviceinstance/redis -n $namespace -o jsonpath='{ .status.conditions[0].reason }'
```
6. Create a function.
```bash
kubectl create -f function/redis-client.yaml -n $namespace
```
7. Create a Binding Usage with **APP_** prefix.
```bash
kubectl create -f function/service-binding-usage.yaml -n $namespace
```
8. Wait until the Function is ready.
```bash
kubeless function ls redis-client --namespace $namespace
```
9. Trigger the Function.
```bash
kubeless function call redis-client --namespace $namespace
```

The information and statistics about the Redis server appear.

>**NOTE:** You can complete steps one to five through the UI.

To perform a clean-up, remove the Namespace:

```bash
kubectl delete ns $namespace
```

## Pluggable SBU

The feature flag **APP_PLUGGABLE_SBU=true** changes Binding Usage Controller logic. The following scenario shows how to test it:

1. Start the application:
```bash
APP_APPLIED_SBU_CONFIG_MAP_NAME=binding-usage-controller-process-sbu-spec APP_LOGGER_LEVEL=debug APP_PLUGGABLE_SBU=true APP_KUBECONFIG_PATH=~/.kube/config go run cmd/controller/main.go
```

2. Register **UsageKind** resources:
```bash
kubectl apply -f usage-kind/deployment.yaml
```

3. Follow the steps under the steps **Use bindings on a Deployment** or **Use bindings on a Function**
