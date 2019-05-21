---
title: Namespaces
type: Details
---

A Namespace is a security and organizational unit which allows you to divide the cluster into smaller units to use for different purposes, such as development and testing.

Namespaces available for users are marked with the `env: "true"` label. The Kyma UI only displays the Namespaces marked with the `env: "true"` label.


## Default Kyma Namespaces

Kyma comes configured with default Namespaces dedicated for system-related purposes. The user cannot modify or remove any of these Namespaces.

- `kyma-system` - This Namespace contains all of the Kyma Core components.
- `kyma-integration` - This Namespace contains all of the Application Connector components responsible for the integration of Kyma and external solutions.
- `kyma-installer` - This Namespace contains all of the Kyma Installer components, objects, and Secrets.
- `istio-system` - This Namespace contains all of the Istio-related components.

## Namespaces for users in Kyma

Kyma comes with three Namespaces ready for you to use.
- `production`
- `qa`
- `stage`

### Create a new Namespace for users

Create a Namespace and mark it with the `env: "true"` label to make it available for Kyma users. Use this command to do that in a single step:

```
$ cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
  labels:
    env: "true"
EOF
```
