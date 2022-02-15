## Overview

Kyma Binding is a component that allows you to inject values from a Secret or ConfigMap to the 
resources such as Deployment, which have a Pod template underneath.

## Prerequisites

- Helm in version 3.0 or higher
- Kubernetes in version 1.16 or higher

## Installation

To install the Kyma Binding component on your cluster, run:

```bash
helm install binding ./charts/binding --wait
```

## Walkthrough

Go through the short tutorial that presents the useage and capabilities of the component.

1. Register a TargetKind. It's a resource that defines how and where environments can be injected.

```bash
kubectl apply -f ./examples/target-kind-deployment.yaml
``` 

2. Check if the TargetKind is registered properly:

```bash
kubectl get targetkinds.bindings.kyma-project.io
```

3. Create a Deployment into which environments will be injected:

```bash
kubectl apply -f ./examples/deployment.yaml
```

4. Port-forward the Pod when it's ready:

```bash
kubectl port-forward svc/env-sample 8080:8080
```

5. Check the browser on ```http://localhost:8080```. You can see that there are no environments injected.

6. Create a Secret:

```bash
kubectl create secret generic secret-with-credentials --from-literal=APP_PASSWORD='super_secret_password' --from-literal=APP_TOKEN='token_to_app'
``` 

7. When the Secret is ready, create a Binding that injects data from the Secret to the Pod:

```bash
kubectl apply -f ./examples/secret-binding.yaml
```

8. Check again the browser. Environments are now injected under the **Password** and **Token** keys.

9. You can also inject environments from a ConfigMap. To check this out, create a ConfigMap and a Binding that injects parameters to our Pod:

```bash
kubectl apply -f ./examples/config-map.yaml
kubectl apply -f ./examples/config-map-binding.yaml
```

10. Check the browser on ```http://localhost:8080```. There are three environments injected.
