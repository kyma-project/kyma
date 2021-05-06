---
title: Permission Controller chart
type: Configuration
---

To configure the Permission Controller chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

The following table lists the configurable parameters of the permission-controller chart and their default values.

| Parameter | Description | Default value |
| --------- | ----------- | ------------- |
| `global.kymaRuntime.namespaceAdminGroup` | Determines the user group for which a RoleBinding to the **kyma-namespace-admin** role is created in all Namespaces except those specified in the `config.namespaceExcludeList` parameter. | `runtimeNamespaceAdmin` |
| `config.namespaceExcludeList` | Comma-separated list of Namespaces in which a RoleBinding to the **kyma-namespace-admin** role is not created for the members of the group specified in the `global.kymaRuntime.namespaceAdminGroup` parameter.|`kyma-system, istio-system, default, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss, compass-system` |
| `config.enableStaticUser`| Determines if a RoleBinding to the **kyma-namespace-admin** role for the static `namespace.admin@kyma.cx` user is created for every Namespace that is not blocked. | `true` |

## Customization examples
You can adjust the default settings of the Permission Controller by applying these overrides to the cluster either before installation, or at runtime:

1. To change the default group, run:

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
  global.kymaRuntime.namespaceAdminGroup: "{CUSTOM-GROUP}"
EOF
```

2. To change the excluded Namespaces and decide whether the `namespace.admin@kyma.cx` static user should be assigned the **kyma-admin** role, run:

```bash
cat <<EOF | kubectl apply -f -
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
  config.namespaceExcludeList: "kyma-system, istio-system, default, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss, compass-system, {USER-DEFINED-NAMESPACE-1}, {USER-DEFINED-NAMESPACE-2}"
  config.enableStaticUser: "{BOOLEAN-VALUE-FOR-NAMESPACE-ADMIN-STATIC-USER}"
EOF
```

3. To change the group mappings run:

```bash
cat <<EOF | kubectl apply -f -
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
  # namespace admins group name
  global.kymaRuntime.namespaceAdminGroup: "runtimeNamespaceAdmin"
  # namespace developers group name
  global.kymaRuntime.developerGroup: "runtimeDeveloper"
  # cluster wide kyma admins group name
  global.kymaRuntime.adminGroup: "runtimeAdmin"
EOF
```
