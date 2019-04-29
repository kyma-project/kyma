---
title: Overview
type: Configuration
---

## Default settings

During the installation process, the Installer applies onto a cluster or Minikube all components defined in the `installer-cr-cluster.yaml.tpl` or `installer-cr.yaml.tpl` templates with their configuration defined in `values.yaml` files. It also imports the configuration overrides defined in the `installer-config-cluster.yaml.tpl` and `installer-config-local.yaml.tpl` templates located under the `installation/resources` subfolder.

> **NOTE:** The installation and configuration templates serve as the basis for creating the corresponding installation (`kyma-config-local.yaml` and `kyma-config-cluster.yaml`) and configuration (`kyma-installer-local.yaml` and `kyma-installer-cluster.yaml`) release artefacts.

Both configuration files contain pre-defined overrides in the form of Secrets and ConfigMaps that the Installer uses during the installation to replace default values specified in `values.yaml` files or to provide required configuration values. While the local configuration file contains hardcoded override values, the cluster configuration file is based on placeholders that are replaced with real values during the installation process.
You can add multiple Secrets and ConfigMaps for one component. If you duplicate the same parameter in a few Secrets or ConfigMaps for a given component but give it different values, the Installer accepts the last value in the file.

>**CAUTION:** Both `values.yaml` and the configuration files contain production settings that are tested and recommended. Note that you modify them on your own risk.

## Installation configuration

Before you start the Kyma installation process, you can customize the default settings in the installation and configuration files.

### Components

During the installation, the Installer creates the Kyma Installer image that contains all component charts located in the `kyma/resources` folder. Both cluster and local installation files contain a full list of these component names and their Namespaces.
Components that are not an integral part of the default Kyma Lite package, are preceded with hashtags (#) which means that the Installer skips them during the installation process.   
You can customize the component installation files by:
- Removing hashtags in front of the component entries to enable a component installation.
- Adding hashtags to disable a component installation.
- Adding new components to the list along with their chart definition in the `kyma/resources` folder. In that case you must create your own [Kyma Installer image](#configuration-use-your-own-kyma-installer-image) as you are adding new chart configuration.

For more details on custom component installation, see [this](#configuration-custom-component-installation) document.

### Overrides

The configuration files provide production-ready settings for the Installer. These include ConfigMaps and Secrets with the overrides for the values hardcoded in the `values.yaml` files of component charts.

You can modify local settings, such as memory limits for a given resource, by changing values in the existing ConfigMaps and Secrets or adding new ones. Every ConfigMap and Secret that the Installer reads must contain an `installer: overrides` label and a `component: {component-name}` label if it refers to a specific component.

[Read more](#configuration-helm-overrides-for-kyma-installation) about the types of overrides and the rules for creating them.

## Runtime configuration

Upon installing Kyma, you can still modify the installation artifacts. However, for these changes to take effect, you must trigger the update process:  

```
kubectl label installation/kyma-installation action=install
```

>**NOTE:** You cannot uninstall a Kyma component that is already installed. If you remove it from any of the installation files or add hashtags in front of its `name` and `namespace` entries, you only disable its further updates.

Apart from modifying the installation artifacts, you can additionally configure the installed components and their behavior in the runtime using ConfigMaps and Secrets.

For an example of such a runtime configuration, see [Helm Broker configuration](/components/helm-broker/#configuration-configuration) in which you add links the bundle repositories to a ConfigMap and label it with the `helm-broker-repo=true` label for the Helm Broker to expose additional Service Classes in the Service Catalog.

## Advanced configuration

All `values.yaml` files in charts and sub-charts contain pre-defined attributes that are:
- Configurable
- Used in chart templates
- Recommended production settings

You can only override values that are included in the `values.yaml` files of a given resource. If you want to extend the available customization options, request adding a new attribute with a default value to the pre-defined list in `values.yaml`. Raise a pull request in which you propose changes in the chart, the new attribute, and its value to be added to `values.yaml`. This way you ensure that you can override these values when needed, without these values being overwritten each time an update or rebase takes place.

>**NOTE:** Avoid modifications of such open-source components as Istio or Service Catalog as such changes can severely impact their future version updates.
