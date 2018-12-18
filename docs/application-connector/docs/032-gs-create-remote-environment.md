---
title: Create a new Remote Environment
type: Getting Started
---

The Application Operator listens for the creation of Remote Environment custom resources. It provisions and de-provisions the necessary deployments for every created Remote Environment (RE).

>**NOTE:** A Remote Environment represents a single connected external solution.

To create a new RE, run this command:

```
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: {RE_NAME}
spec:
  description: {RE_DESCRIPTION}
  labels:
    region: us
    kind: production
EOF
```

## Check the RE status

To check the status of the created RE and show the output in the `yaml` format, run this command:
```
kubectl get re {RE_NAME} -o yaml
```

A successful response returns the RemoteEnvironment custom resource with the specified name. The custom resource has the **status** section added.
This is an example response: 

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  clusterName: ""
  creationTimestamp: 2018-11-22T13:53:20Z
  generation: 1
  name: test1
  namespace: ""
  resourceVersion: "30728"
  selfLink: /apis/applicationconnector.kyma-project.io/v1alpha1/remoteenvironments/test1
  uid: f8ca5595-ee5d-11e8-acb2-000d3a443243
spec:
  accessLabel: {RE_NAME}
  description: {RE_DESCRIPTION}
  labels: {}
  services: []
status:
  installationStatus:
    description: Installation complete
    status: DEPLOYED
```
