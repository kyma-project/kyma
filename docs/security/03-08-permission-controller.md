---
title: Permission Controller
type: Details
---

The Permission Controller is a Kubernetes controller which listens for new Namespaces and creates RoleBindings for the users of specified groups to the **kyma-admin** role within these Namespaces. The Controller uses a blacklist mechanism, which defines the Namespaces in which the users of the defined groups are not assigned the **kyma-admin** role. 

When the Controller is deployed in a cluster, it checks all existing Namespaces and groups and assigns the roles accordingly.

To support the functionality of the controller, the users of the selected groups have a ClusterRoleBinding to the **kyma-namespace-admin-essentials** role, which allows them to view cluster resources through the Console UI, list cluster resources, and create Namespaces.

By default, the controller binds users of the **namespace-admins** group to the **kyma-admin** role in the Namespaces they create. Additionally, the default configuration creates a ClusterRoleBinding for the users of the **namespace-admins** group to the **kyma-namespace-admin-essentials** role and creates a RoleBinding for the static `namespace.admin@kyma.cx` user to the **kyma-admin** role in every Namespace that is not blacklisted. 

You can adjust the default settings of the Permission Controller by applying these overrides to the cluster either before installation, or at runtime: 

>**TIP:** To learn more about the adjustable parameters of the Permission Controller, read this document. 

1. To change the default groups, run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: namespace-admin-groups-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.users.namespaceAdmin.groups: "namespace-admins, {CUSTOM-GROUP-1}, {CUSTOM-GROUP-2}"
EOF
```

2. To change the blacklisted Namespaces and decide whether the `namespace.admin@kyma.cx` static user should be assigned the **kyma-admin** role, run: 

```bash
apiVersion: v1
kind: ConfigMap
metadata:
  name: permission-controller-overrides
  namespace: kyma-installer
  labels:
    component: permission-controller
    installer: overrides
    kyma-project.io/installation: ""
data:
  config.namespaceBlacklist: "kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss, {USER-DEFINED-NAMESPACE-1}, {USER-DEFINED-NAMESPACE-2}"
  config.enableStaticUser: "{BOOLEAN-VALUE-FOR-NAMESPACE-ADMIN-STATIC-USER"
EOF
```
