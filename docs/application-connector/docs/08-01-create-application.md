---
title: Create a new Application
type: Tutorials
---

The Application Operator listens for the creation of Application custom resources. It provisions and de-provisions the necessary deployments for every created Application (App).

>**NOTE:** An App represents a single connected external solution.

To create a new App, run this command:

```
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
  name: {APP_NAME}
spec:
  description: {APP_DESCRIPTION}
  labels:
    region: us
    kind: production
EOF
```

## Check the App status

To check the status of the created App and show the output in the `yaml` format, run this command:
```
kubectl get app {APP_NAME} -o yaml
```

A successful response returns the Application custom resource with the specified name. The custom resource has the **status** section added.
This is an example response:

```
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
  labels: {}
  services: []
status:
  installationStatus:
    description: Installation complete
    status: DEPLOYED
```
