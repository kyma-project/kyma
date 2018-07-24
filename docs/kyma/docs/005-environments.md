---
title: Environments
type: Details
---

An Environment is a custom Kyma security and organizational unit based on the concept of Kubernetes [Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/). Kyma Environments allow you to divide the cluster
into smaller units to use for different purposes, such as development and testing.

Kyma Environment is a user-created Namespace marked with the `env: "true"` label. The Kyma UI only displays the Namespaces marked with the `env: "true"` label.


## Default Kyma Namespaces

Kyma comes configured with default Namespaces dedicated for system-related purposes. The user cannot modify or remove any of these Namespaces.

- `kyma-system` - This Namespace contains all of the Kyma Core components.
- `kyma-integration` - This Namespace contains all of the Application Connector components responsible for the integration of Kyma and external solutions.
- `kyma-installer` - This Namespace contains all of the Kyma installer components, objects, and Secrets.
- `istio-system` - This Namespace contains all of the Istio-related components.

## Environments in Kyma

Kyma comes with three Environments ready for you to use. These environments are:

- `production`
- `qa`
- `stage`

## Create a new Environment

To create a new Environment, create a Namespace and mark it with the `env: "true"` label. Use this command to do that in a single step:

```
$ cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Namespace
metadata:
  name: my-environment
  labels:
    env: "true"
EOF
```

Initially, the system deploys two template roles: `kyma-reader-role` and `kyma-admin-role`. The controller finds the template roles by filtering available roles in the namespace `kyma-system` by the label `env: "true"`. The controller copies these roles into the Environment.
