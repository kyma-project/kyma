## Overview

Kyma Binding is a component that allows injecting values from Secret/ConfigMap to the 
resources which have a Pod template underneath. (e.g. Deployment)

## Installation

To install the Kyma Binding component to your cluster use Helm in version >= 3.0

```bash
helm install binding ./charts/binding --wait
```

## Walkthrough

Below there is a short tutorial presenting the use and capabilities of the component.

Our first step is to register TargetKind, a resource which defines how and where environments could be injected:

```bash
kubectl apply -f ./examples/target-kind-deployment.yaml
``` 

check if TargetKind was registered properly:

```bash
kubectl get targetkinds.bindings.kyma-project.io
```

create Deployment into which environments will be injected:

```bash
kubectl apply -f ./examples/deployment.yaml
```

when Pod will be ready you can port-forward it:

```bash
kubectl port-forward svc/env-sample 8080:8080
```

and check the browser on ```http://localhost:8080``` web content with not injected environments.

The next step is to create secret:

```bash
kubectl create secret generic secret-with-credentials --from-literal=APP_PASSWORD='super_secret_password' --from-literal=APP_TOKEN='token_to_app'
``` 

when Secret will be ready, create Binding which inject data from Secret to our Pod:

```bash
kubectl apply -f ./examples/secret-binding.yaml
```

now check again web browser, environments should be injected under `Password` and `Token` keys.

In addition to Secrets, environments can be injected from ConfigMap, to test this create ConfigMap and 
Binding injecting parameters to our Pod:

```bash
kubectl apply -f ./examples/config-map.yaml
kubectl apply -f ./examples/config-map-binding.yaml
```

check web browser on ```http://localhost:8080```, three environments should be injected.
