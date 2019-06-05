---
title: Overview
type: Configuration
---

## Introduction

You can configure the Kyma installation by:
  - Customizing the list of the components to install.
  - Providing overrides that change the configuration values used by one or more components.

The list of components to install is defined in the [Installation](#custom-resource-installation) custom resource (CR).
The overrides are delivered as ConfigMaps or Secrets defined by the user before triggering the installation.
The Installer reads the configuration from the Installation CR and the overrides and applies it in the installation process.


## Default settings

The default settings for the cluster and local installation are defined in different files.

<div tabs>
  <details>
  <summary>
  Local installation
  </summary>

  For the list of all components available to install see the `installer-cr.yaml.tpl` file.
  For the list of the default installation overrides see the `installer-config-local.yaml.tpl` file.
  Other configuration values are defined directly in the configuration of the respective components.
  >**CAUTION:** The default configuration uses tested and recommended settings. Change them at your own risk.
  </details>
  <details>
  <summary>
  Cluster installation
  </summary>

  The default installation flow uses a Kyma release.
  All components available in a given release are listed in the  `kyma-installer-cluster.yaml`, which is one of the release artifacts.
  Any required overrides are described in the [cluster installation guide](#installation-install-kyma-on-a-cluster).
 Other settings are defined directly in the configuration of the components released with the given Kyma version.
  </details>
</div>

## Installation configuration

Before you start the Kyma installation process, you can customize the default settings.

### Components

One of the released Kyma artifacts is the Kyma-Installer, a Docker image that combines the Installer executable with charts for all components available in the release.
The Kyma-Installer can install only the components contained in its image.
The Installation CR specifies which components of the available components are installed.
The component list in the Installation CR has the components that are not an integral part of the default Kyma Lite package commented out with a hash character (#). The Installer doesn't install these components.
You can customize the list of components by:
- Uncommenting a component entry to install the component.
- Commenting out a component entry using the hash character (#) to skip the installation of that component.
- Adding new components along with their chart definitions to the list. If you do that, you must build your own [Kyma Installer image](#installation-use-your-own-kyma-installer-image) as you are adding a new component to Kyma.

For more details on custom component installation, see [this](#configuration-custom-component-installation) document.

### Overrides

The common overrides that affect the entire installation, are described in the installation guides.
Other overrides are component-specific.
To learn more about the configuration options available for a specific component, see the **Configuration** section of the component's documentation.
Once you know the name and possible set of values for a configuration option, define an override for it by creating a ConfigMap or a Secret. You can also extend an existing one.
[Read more](#configuration-helm-overrides-for-kyma-installation) about the types of overrides and the rules for creating them.
>**CAUTION:** An override must exist in a cluster before the installation is started, otherwise the Installer is not be able to apply it correctly.

> **CAUTION:** In documentation for each Kyma component, you can find configuration documents that list configurable parameters from the `values.yaml` files for a given component chart or sub-chart. Override values only for the parameters exposed in those configuration documents.

## Runtime configuration

Changing static configuration of the Kyma after installation is generally not supported.
Some components may support custom runtime configuration changes though.
For an example of such a runtime configuration, see [Helm Broker configuration](/components/helm-broker/#configuration-configuration) in which you add links to the bundle repositories to a ConfigMap and label it with the `helm-broker-repo=true` label for the Helm Broker to expose additional Service Classes in the Service Catalog.

Another solution for changing components configuration after installation is to alter its overrides and update the component using Kyma's [Update procedure](#installation-update-kyma).
Support for this is limited only to component-specific configuration options and depends on the component. Refer to the component's documentation for details.


## Advanced configuration

All `values.yaml` files in charts and sub-charts contain pre-defined attributes that are:
- Configurable
- Used in chart templates
- Recommended production settings

You can only override values that are included in the `values.yaml` files of a given resource. If you want to extend the available customization options, request adding a new attribute with a default value to the pre-defined list in `values.yaml`. Raise a pull request in which you propose changes in the chart, the new attribute, and its value to be added to `values.yaml`. This way you ensure that you can override these values when needed, without these values being overwritten each time an update or rebase takes place.

>**NOTE:** Avoid modifications of such open-source components as Istio or Service Catalog as such changes can severely impact their future version updates.
