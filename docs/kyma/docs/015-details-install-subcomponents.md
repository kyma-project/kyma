---
title: Install subcomponents
type: Details
---

It is up to you to decide which subcomponents you install as part of the `core` release. By default, most of the core subcomponents are enabled. If you want to install only specific subcomponents, follow the steps that you need to perform before the local and cluster installation.

## Install subcomponents locally

To specify whether to install a given core subcomponent on Minikube, use the `manage-component.sh` script before you trigger the Kyma installation. The script consumes two parameters:

- the name of the core subcomponent
- a Boolean value that determines whether to install the subcomponent (`true`) or not (`false`)

Example:

To enable the Azure Broker subcomponent, run the following command:
```
scripts/manage-component.sh azure-broker true
```

Alternatively, to disable the Azure Broker subcomponent, run this command:
```
scripts/manage-component.sh azure-broker false
```

## Install subcomponents on a cluster

Install subcomponents on a cluster based on Helm conditions described in the `requirements.yaml` file. Read more about the fields in the `requirements.yaml` file [here](https://github.com/helm/helm/blob/release-2.10/docs/charts.md#tags-and-condition-fields-in-requirementsyaml).

To specify whether to install a given core subcomponent, provide override values before you trigger the installation.

Example:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: kyma-sub-components
  namespace: kyma-installer
  labels:
    installer: overrides
data:
  azure-broker.enabled: "true"
```

>**NOTE:** Some subcomponents can require additional configuration to work properly.

## Specify subcomponents versions

Versions of the Kyma components are specified in the `values.yaml` file in charts. Two properties, `version` and `dir`, describe each component version. The first one defines the actual docker image tag. The second property describes the directory under which the tagged image is pushed. It is optional and is followed by a forward slash (/).

Possible values of the `dir` property:
- `pr/` contains images built from the pull request
- `develop/` contains images built from the `master` branch
- `rc/` contains images built for a pre-release
- `` (empty) contains images built for a release

To override subcomponents versions during Kyma startup, create the `versions-overrides.env` file in the `installation` directory.

The example overrides the `Environments` component and sets the image version to `0.0.1`, based on the version from the `develop` directory.

Example:

```
global.environments.dir=develop/
global.environments.version=0.0.1
```
