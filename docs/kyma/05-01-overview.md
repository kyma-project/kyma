---
title: Overview
type: Configuration
---

## Introduction

Users can configure Kyma installation in two ways:
  - Customize the list of the components to install
  - Provide overrides that change configuration values used by one or more components

The list of components to install is defined in the **Installation** Custom Resource.
Overrides are defined as ConfigMap or Secret objects defined by the user before the installation starts.
The Installer is the Kyma component that is responsible for reading and applying the configuration.


## Default settings

Default configuration is defined differently for cluster and local installations.

<div tabs>
  <details>
  <summary>
  Local installation
  </summary>

  For local installation, the list of available components is stored in Kyma project sources in the file **kyma/installation/resources/installer-cr.yaml.tpl**
  Default installation overrides can be found in the file **kyma/installation/resources/installer-config-local.yaml.tpl**
  All the other configuration values are defined directly in components in **/resources** subdirectory of Kyma project.
  >**CAUTION:** The configuration files contain tested and recommended settings. Note that you modify them on your own risk.
  </details>
  <details>
  <summary>
  Cluster installation
  </summary>

  Cluster installations are based on a released Kyma version.
  You can find the list of available components in the release artifact **kyma-installer-cluster.yaml**.
  The necessary overrides, if any, are described in relevant [installation procedure](#installation-install-kyma-on-a-cluster).
  All the other configuration values are defined directly in components released with the Kyma version.
  </details>
</div>

## Installation configuration

Before you start the Kyma installation process, you can customize the default settings.

### Components

One of the released Kyma artifacts is the Kyma-Installer, a Docker image that combines the Installer executable with charts for all components available in the release.
The Kyma-Installer can install only the components contained in it's image.
**Installation** Custom Resource specifies which components of all available ones should be actually installed.
Components that are not an integral part of the default Kyma Lite package are commented out with a hash character (#) and skipped during installation.
You can customize list of components used for installation by:
- Uncommenting component entry to enable a component installation.
- Commenting out an entry using hash chacter to disable a component installation.
- Adding new components to the list along with their chart definition. In that case you must create your own [Kyma-Installer image](#installation-use-your-own-kyma-installer-image) as you are adding new component to Kyma.

For more details on custom component installation, see [this](#configuration-custom-component-installation) document.

### Overrides

Common overrides that affect entire installation, for example `global.domainName`, are already described in the [installation procedure](#installation-overview).
Other overrides are component-specific.
To learn about configuration options for a given component, refer to **Configuration** section of the component's documentation.
Once you know the name and possible set of values for a configuration option, define an override for it by creating a ConfigMap or a Secret. You can also extend an existing one.
[Read more](#configuration-helm-overrides-for-kyma-installation) about the types of overrides and the rules for creating them.
>**CAUTION:** An override must exist in a cluster before the installation is started, otherwise the Installer is not be able to apply it correctly.

> **CAUTION:** In documentation for each Kyma component, you can find configuration documents that list configurable parameters from the `values.yaml` files for a given component chart or sub-chart. Override values only for the parameters exposed in those configuration documents.

## Runtime configuration

Changing static configuration of the Kyma after installation is generally not supported.
Some components may support custom runtime configuration changes though.
For an example of such a runtime configuration, see [Helm Broker configuration](/components/helm-broker/#configuration-configuration) in which you add links to the bundle repositories to a ConfigMap and label it with the `helm-broker-repo=true` label for the Helm Broker to expose additional Service Classes in the Service Catalog.

Another solution for changing components configuration after installation is to alter it's overrides and update the component using Kyma's [Update procedure](#installation-update-kyma).
Support for this is limited only to component-specific configuration options and depends on the component. Refer to the component's documentation for details.


## Advanced configuration

All `values.yaml` files in charts and sub-charts contain pre-defined attributes that are:
- Configurable
- Used in chart templates
- Recommended production settings

You can only override values that are included in the `values.yaml` files of a given resource. If you want to extend the available customization options, request adding a new attribute with a default value to the pre-defined list in `values.yaml`. Raise a pull request in which you propose changes in the chart, the new attribute, and its value to be added to `values.yaml`. This way you ensure that you can override these values when needed, without these values being overwritten each time an update or rebase takes place.

>**NOTE:** Avoid modifications of such open-source components as Istio or Service Catalog as such changes can severely impact their future version updates.
