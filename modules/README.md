# Modules

## Overview

Modules are the next generation of components in Kyma that are available for local and cluster installation.

Modules are no longer represented by a single helm-chart, but instead are bundled and released within channels through a `ModuleTemplate`, a unique link of a module, and its desired state of charts and configuration, and a channel.

This directory includes public `ModuleTemplates` for all channels available for consumption within Kyma, wherein each channel is represented by a dedicated kustomization.

## Development

To apply all `ModuleTemplates` for your cluster, use the kustomization available in this folder:

```
kubectl apply -k modules
```

Then you can enable available Modules either with support of Busola through the `Kyma` Custom Resource, or you can manually edit the `Kyma CR` for the same effect.

## Adding a new Module

The only type of Module currently accepted is `ModuleTemplate` in version `v1alpha1`.

When submitting a new module, make sure to add it to the proper `kustomization.yaml` to also integrate it with our rendering.

## Disclaimer

Since the Kyma Module Eco-System is still in an early stage, expect changes to this folder and its offered kustomization.