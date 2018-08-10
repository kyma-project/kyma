---
title: Installation
type: Custom Resource
---

The `installations.installer.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to control the Kyma Installer, a proprietary solution based on the
[Kubernetes operator](https://coreos.com/operators/) principles. To get the up-to-date CRD and show the output in the `yaml` format, run this command:  

```
kubectl get crd installations.installer.kyma.cx -n kyma-installer -o yaml
```

## Sample Custom Resource

This is a sample CR that controls the Kyma installer. This example has the **action** label set to `install`, which means that it triggers the installation of Kyma.

```
apiVersion: "installer.kyma.cx/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  labels:
    action: install
  finalizers:
    - finalizer.installer.kyma.cx
spec:
  version: "1.0.0"
  url: "https://sample.url.com/kyma_release.tar.gz"
```

This table analyses the elements of the sample CR and the information it contains:

| Field   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **apiVersion** | **YES** | Defined basing on the **group** and **version** fields of the CRD **spec** section. |
| **kind** | **YES** | Defined basing on the **names: kind** field of the CRD **spec** section. |
| **metadata.name** | **YES** | Specifies the name of the CR. |
| **metadata.labels.action** | **YES** | Defines the behavior of the Kyma installer. Available options are `install` and `uninstall`. |
| **metadata.finalizers** | **NO** | Protects the CR from deletion. Read [this](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers) Kubernetes document to learn more about finalizers. |
| **spec.version** | **NO** | When manually installing Kyma on a cluster, specify any valid [SemVer](https://semver.org/) notation string. |
| **spec.url** | **YES** | Specifies the location of the Kyma sources `tar.gz` package. For example, for the `master` branch of Kyma, the address is `https://github.com/kyma-project/kyma/archive/master.tar.gz` |
