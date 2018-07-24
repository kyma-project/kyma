# Remote environment binding test

## Overview

This folder contains the test which checks if the BindingUsage resource allows the Pod to call the fake Gateway.

## Details

The testing scenario has the following steps:
1. Setup: create RemoteEnvironment, EnvironmentMapping, deployments (`fake-gateway`, `gateway-client`) and set-up Istio Deniers and Rules
2. Provision a ServiceClass
3. Perform binding
4. Add a BindingUsage
5. The `gateway-client` can call the `fake-gateway` (an environment variable with Gateway URL is injected, Istio allows the call)
6. Remove BindingUsage
5. The `gateway-client` cannot call the `fake-gateway` (there is no environment variable with Gateway URL, Istio blocks the call)

The `gateway-client` stores values read by the test in a ConfigMap.

## Usage

This section explains how to run the test on the cluster.

### Setup

Go to the project root directory.

Build testing Docker image:
```bash
./remote-environment/contrib/build.sh
```

Create service accounts and roles:
```bash
kubectl apply -f remote-environment/contrib/rbac.yaml
```

### Run the test

Create a testing Pod:
```bash
kubectl apply -f remote-environment/contrib/pod.yaml
``` 

### Watch resources
The test creates and updates Kubernetes resources in the `acceptance-test` Namespace. You can observe the test's progress using the following command:
```bash
kubectl get configmap,po,svc,servicebindingusage,servicebinding -n acceptance-test
```

You can see the ConfigMap with the values saved by the `gateway-client`:
```bash
kubectl get configmap -n acceptance-test -o yaml
``` 

### Cleanup
Clean up all test resources:

```bash
kubectl delete ns acceptance-test
kubectl delete po -n kyma-system re-acceptance-test
```
