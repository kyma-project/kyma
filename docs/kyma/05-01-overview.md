---
title: Overview
type: Configuration
---

## Introduction

Users can configure the way Kyma is installed in two ways:
  - Customizing the list of the components to install
  - Providing overrides that allow to change configuration values used by one or more components

The list of components to install is defined in the **Installation** Custom Resource.
Overrides are defined as ConfigMap or Secret objects defined by the user before the installation starts.
The Installer is the Kyma component that is responsible for reading and applying the configuration.


## Default settings

Default configuration is defined differently for cluster and local installations.

### Cluster installation
Cluster installations are based on a released Kyma version.
You can find the list of available components in the release artifact **kyma-installer-cluster.yaml**.
The necessary overrides, if any, are described in relevant [installation procedure](https://kyma.project.io/).
All the other configuration values are defined directly in components released with the Kyma version.

### Local installation
For local installation the list of available components is stored in Kyma project sources in the file **kyma/installation/resources/installer-cr.yaml.tpl**
Default installation overrides can be found in the file **kyma/installation/resources/installer-config-local.yaml.tpl**
>**CAUTION:** The configuration files contain tested and recommended settings. Note that you modify them on your own risk.
All the other configuration values are defined directly in components in **/resources** subdirectory of Kyma project.

## Installation configuration

Before you start the Kyma installation process, you can customize the default settings.

### Components

One of the released artifacts is the Kyma-Installer, a Docker image that combines the Installer executable with charts for all components available in the release.
The Kyma-Installer can install only the components from it's image.
**Installation** Custom Resource specifies which components of all available ones should be actually installed.
Components that are not an integral part of the default Kyma Lite package are commented out with a hash character (#).
It means that the Installer skips them during installation.
You can customize component installation files by:
- Uncommenting component entries to enable a component installation.
- Adding hashtags to disable a component installation.
- Adding new components to the list along with their chart definition. In that case you must create your own [Kyma-Installer image](#installation-use-your-own-kyma-installer-image) as you are adding new component to Kyma.

For more details on custom component installation, see [this](#configuration-custom-component-installation) document.

### Overrides

#### Cluster installations
Common overrides that affect entire installation, e.g. domainName, are already described in the cluster [installation procedure](https://kyma.project.io/).
Most other overrides are component-specific.
To learn available configuration options for a given component, refer to **Configuration** section of the component documentation.
Once you know the name and possible set of values for a configuration option, you can define an override for it by extending an existing ConfigMap/Secret object, or creating a new one.
[Read more](#configuration-helm-overrides-for-kyma-installation) about the types of overrides and the rules for creating them.
>**CAUTION:** An override must exist in a cluster before the installation is started, otherwise the Installer is not able to apply it.


## Runtime configuration

Once Kyma installation completes, you can still modify the installation artifacts. However, for these changes to take effect, you must trigger the update process:

```
kubectl label installation/kyma-installation action=install
```

>**NOTE:** You cannot uninstall a Kyma component that is already installed. If you remove it from any of the installation files or add hashtags in front of its `name` and `namespace` entries, you only disable its further updates.

Apart from modifying the installation artifacts, you can use ConfigMaps and Secrets to configure the installed components and their behavior in the runtime.

For an example of such a runtime configuration, see [Helm Broker configuration](/components/helm-broker/#configuration-configuration) in which you add links to the bundle repositories to a ConfigMap and label it with the `helm-broker-repo=true` label for the Helm Broker to expose additional Service Classes in the Service Catalog.

## Advanced configuration

All `values.yaml` files in charts and sub-charts contain pre-defined attributes that are:
- Configurable
- Used in chart templates
- Recommended production settings

You can only override values that are included in the `values.yaml` files of a given resource. If you want to extend the available customization options, request adding a new attribute with a default value to the pre-defined list in `values.yaml`. Raise a pull request in which you propose changes in the chart, the new attribute, and its value to be added to `values.yaml`. This way you ensure that you can override these values when needed, without these values being overwritten each time an update or rebase takes place.

>**NOTE:** Avoid modifications of such open-source components as Istio or Service Catalog as such changes can severely impact their future version updates.
