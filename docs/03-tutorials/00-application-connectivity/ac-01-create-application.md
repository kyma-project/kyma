---
title: Create a new Application
---

>**NOTE:** An Application represents a single connected external solution.

## Prerequisites

Before you start, export the name of your application as an environment variable:

```bash
export APP_NAME={YOUR_APP_NAME}
```

>**NOTE:** You need to enable Istio sidecar proxy injection, which is disabled on the Kyma clusters by default. To do this, follow the [Enable Istio Sidecar Injection](../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).

## Create an Application

To create a new Application, run this command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
  name: $APP_NAME
spec:
  description: Application description
  labels:
    region: us
    kind: production
EOF
```

## Get the Application data

To get the data of the created Application and show the output in the `yaml` format, run this command:
```bash
kubectl get app $APP_NAME -o yaml
```

A successful response returns the Application custom resource with the specified name. 
This is an example response:

```yaml
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
  clusterName: ""
  creationTimestamp: 2018-11-22T13:53:20Z
  generation: 1
  name: test1
  namespace: ""
  resourceVersion: "30728"
  selfLink: /apis/applicationconnector.kyma-project.io/v1alpha1/applications/test1
  uid: f8ca5595-ee5d-11e8-acb2-000d3a443243
spec:
  accessLabel: {APP_NAME}
  description: {APP_DESCRIPTION}
  labels:
    region: "us"
    kind: "production"
  services: []
```
