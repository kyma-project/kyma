# Service Catalog Addons

## Overview

Service Catalog Addons provides Kyma addons to the [Service Catalog](../service-catalog/README.md).

These addons consist of new views related to the Service Catalog and the Binding Usage Controller which extends the Service Catalog with the additional logic. 

### Views

The Service Catalog Addons provides the following views to the Kyma Console.

* catalog-ui
* instances-ui
* brokers-ui

### Binding Usage Controller

The Binding Usage Controller provides a capability to inject Secrets to the given applications. It introduces two CRs to achieve the injection:

* ServiceBindingUsage
* UsageKind

The Binding Usage Controller chart provides two default UsageKinds to Kyma:

* function-usage-kind.yaml
* deployment-usage-kind.yaml