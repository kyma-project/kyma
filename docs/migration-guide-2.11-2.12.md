---
title: Migration Guide 2.11-2.12
---

Once you upgrade to Kyma 2.12 perform the manual steps described in the Migration Guide.

## Telemetry

### Annotation-based scarping of metrics for system components

In preparation for the upcoming modularization and having a reduced set of dependencies on a module, Kyma switched to an annotation-based scraping of metrics for system components. With that, the ServiceMonitors of the components must be cleaned up. When you upgrade from Kyma 2.11 to 2.12, either run the script [cleanup.sh](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-cleanup-servicemonitors.sh) or run the commands from the script manually.

## Service Management

### PodPreset cleanup

The PodPreset component was deprecated in [Kyma 2.4](https://kyma-project.io/blog/2022/6/30/release-notes-24#pod-preset-deprecation-note) and removed from [Kyma 2.10](https://github.com/kyma-project/kyma/pull/16647). Run the [cleanup script](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-cleanup-podpreset.bash) to remove any residual resources related to PodPresets. Follow [the guide](https://kyma-project.io/blog/2022/6/30/release-notes-24#pod-preset-deprecation-note) to transform the usage of `Secrets` from Kyma `Functions` manually.

## Application Connectivity

### Removal of the `compass-system` Namespace

With Kyma 2.12, Compass Runtime Agent will deploy in the `kyma-system` Namespace instead of `compass-system`. After upgrading Kyma to the 2.12 version, you must execute [this script](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-SKR-delete-compass-system-namespace.sh) that removes the `compass-system` Namespace and patches the CompassConnection custom resource.
