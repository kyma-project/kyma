---
title: Installation
type: Custom Resource
---

The `installations.installer.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to control the Kyma Installer, a proprietary solution based on the
[Kubernetes operator](https://coreos.com/operators/) principles. To get the up-to-date CRD and show the output in the `yaml` format, run this command:  

```
kubectl get crd installations.installer.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR that controls the Kyma Installer. This example has the **action** label set to `install`, which means that it triggers the installation of Kyma. The  **name** and **namespace**  fields in the `components` array define which components you install and Namespaces in which you install them.

>**NOTE:** See the `installer-cr.yaml.tpl` file in the `/installation/resources` directory for the complete list of Kyma components.

```
apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  labels:
    action: install
  finalizers:
    - finalizer.installer.kyma-project.io
spec:
  version: "1.0.0"
  url: "https://sample.url.com/kyma_release.tar.gz"
  components:
    - name: "cluster-essentials"
      namespace: "kyma-system"
    - name: "istio"
      namespace: "istio-system"
    - name: "prometheus-operator"
      namespace: "kyma-system"
    - name: "provision-bundles"
    - name: "dex"
      namespace: "kyma-system"
    - name: "core"
      namespace: "kyma-system"
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|----------|:-------------:|:------|
| **metadata.name** | **YES** | Specifies the name of the CR. |
| **metadata.labels.action** | **YES** | Defines the behavior of the Kyma Installer. Available options are `install` and `uninstall`. |
| **metadata.finalizers** | **NO** | Protects the CR from deletion. Read [this](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers) Kubernetes document to learn more about finalizers. |
| **spec.version** | **NO** | When manually installing Kyma on a cluster, specify any valid [SemVer](https://semver.org/) notation string. |
| **spec.url** | **YES** | Specifies the location of the Kyma sources `tar.gz` package. For example, for the `master` branch of Kyma, the address is `https://github.com/kyma-project/kyma/archive/master.tar.gz` |
| **spec.components** | **YES** | Lists which components of Helm chart components to install or update. |
| **spec.components.name** | **YES** | Specifies the name of the component which is the same as the name of the component subdirectory in the `resources` directory. |
| **spec.components.namespace** | **YES** | Defines the Namespace in which you want the Installer to install, or update the component. |
| **spec.components.release** | **NO** | Provides the name of the Helm release. The default parameter is the component name. |

## Additional information

The Kyma Installer adds the **status** section which describes the status of Kyma installation. This table lists the fields of the **status** section.

| Field   |      Mandatory      |  Description |
|----------|:-------------:|:------|
| **status.state** | **YES** | Describes the installation state. Takes one of four values. |
| **status.description** | **YES** | Describes the installation step the installer performs at the moment. |
| **status.errorLog** | **YES** | Lists all errors that happen during the installation. |
| **status.errorLog.component** | **YES** | Specifies the name of the component that causes the error. |
| **status.errorLog.log** | **YES** | Provides a description of the error. |
| **status.errorLog.occurrences** | **YES** | Specifies the number of subsequent occurrences of the error. |

The **status.state** field uses one of the following four values to describe the installation state:

|   State   |  Description |
|----------|:-------------|
| **Installed** | Installation successful. |
| **Uninstalled** | Uninstallation successful. |
| **InProgress** | The Installer is still installing or uninstalling Kyma. No errors logged. |
| **Error** | The Installer encountered a problem but it continues to try to process the resource. |

## Related resources and components

These components use this CR:

| Component   |   Description |
|----------|:------|
| Installer  |  The CR triggers the Installer to install, update or delete of the specified components. |
