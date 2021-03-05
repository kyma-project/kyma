---
title: Overview
type: Configuration
---

You can configure the Kyma deployment by:
  - Customizing the list of the components to deploy.
  - Providing overrides that change the configuration values used by one or more components.

The list of components to deploy is defined in the [Deployment](#custom-resource-installation) custom resource (CR).
The overrides are delivered as a `.yaml` file defined by the user before triggering the deployment.
The Kyma Installer reads the configuration from the Installation CR and the overrides and applies it in the deployment process.


## Default settings

The default settings for the cluster and local deployment are defined in different files.

<div tabs name="default-settings" group="configuration">
  <details>
  <summary label="local-installation">
  Local installation
  </summary>

  For the list of all components available to deploy see the `components.yaml` file.
  For the list of the default deployment overrides see the `overrides.yaml` file.
  Other configuration values are defined directly in the configuration of the respective components.

  >**CAUTION:** The default configuration uses tested and recommended settings. Change them at your own risk.
  </details>
  <details>
  <summary label="cluster-deployment">
  Cluster deployment
  </summary>

  The default deployment flow uses a Kyma release.
  All components available in a given release are listed in the  `overrides.yaml`, which is one of the release artifacts.
  Any required overrides are described in the [cluster deployment guide](#installation-install-kyma-on-a-cluster).
  Other settings are defined directly in the configuration of the components released with the given Kyma version.
  </details>
</div>

## Deployment configuration

Before you start the Kyma deployment process, you can customize the default settings.

### Components

Kyma CLI takes the components directly from the released Kyma sources or sources that are checked out locally.
Kyma CLI can deploy the components that are released or that are available locally in case of a deployment from the local sources.
`components.yaml` file specifies the list of components to deploy. You can use your own `components.yaml` file if necessary.
You can customize the list of components by:
- Uncommenting a component entry to deploy the component.
- Commenting out a component entry using the hash character (#) to skip the deployment of that component.
- Adding new components along with their chart definitions to the list. If you do that, you must build your own [Kyma Installer image](#installation-use-your-own-kyma-installer-image) as you are adding a new component to Kyma.

For more details on custom component deployment, see the [configuration document](#configuration-custom-component-installation).

### Overrides

The common overrides that affect the entire deployment, are described in the deployment guides.
Other overrides are component-specific.
To learn more about the configuration options available for a specific component, see the **Configuration** section of the component's documentation.

>**CAUTION:** Override only values for those parameters from `values.yaml` files that are exposed in configuration documents for a given component.

[Read more](#configuration-helm-overrides-for-kyma-installation) about the types of overrides and the rules for creating them.

>**CAUTION:** To pass overrides, use the `overrides.yaml` file. You can also pass them directly by using the command `kyma deploy --value`.

## Advanced configuration

All `values.yaml` files in charts and sub-charts contain pre-defined attributes that are:
- Configurable
- Used in chart templates
- Recommended production settings

You can only override values that are included in the `values.yaml` files of a given resource. If you want to extend the available customization options, request adding a new attribute with a default value to the pre-defined list in `values.yaml`. Raise a pull request in which you propose changes in the chart, the new attribute, and its value to be added to `values.yaml`. This way you ensure that you can override these values when needed, without these values being overwritten each time an update or rebase takes place.

>**NOTE:** Avoid modifications of such open-source components as Istio or Service Catalog as such changes can severely impact their future version updates.
