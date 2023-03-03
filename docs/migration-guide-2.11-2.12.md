---
title: Migration Guide 2.11-2.12
---

Once you upgrade to Kyma 2.12, perform the manual steps described in the Migration Guide.

## Telemetry

### Annotation-based scraping of metrics for system components

In preparation for the upcoming modularization and having a reduced set of dependencies on a module, Kyma switched to an annotation-based scraping of metrics for system components. With that, the ServiceMonitors of the components must be cleaned up. When you upgrade from Kyma 2.11 to 2.12, either run the script [cleanup.sh](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-cleanup-servicemonitors.sh) or run the commands from the script manually.

### Remove Init Container from telemetry-operator deployment

Creating the ValidatingWebhookConfiguration and the related CA bundle was moved from an init container to the operator itself. Run the [cleanup script](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-cleanup-init-container.sh) to remove the old init container.

## Application Connectivity

### Removal of the `compass-system` Namespace

With Kyma 2.12, Compass Runtime Agent will deploy in the `kyma-system` Namespace instead of `compass-system`. After upgrading Kyma to the 2.12 version, you must execute [this script](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.11-2.12-delete-compass-system-namespace.sh) that removes the `compass-system` Namespace and patches the CompassConnection custom resource.
