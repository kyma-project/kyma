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

* [ServiceBindingUsage](../../../../docs/service-catalog/docs/040-cr-service-binding-usage.md)
* [UsageKind](../../../../docs/service-catalog/docs/041-cr-usage-kind.md)

For more information on these CRs, go to the {name} directory.

The Binding Usage Controller chart provides two default UsageKinds to Kyma:

* [function](charts/binding-usage-controller/templates/function-usage-kind.yaml)
* [deployment](charts/binding-usage-controller/templates/deployment-usage-kind.yaml)

For more detailed information, go to the [Binding Usage Controller](https://github.com/kyma-project/kyma/tree/master/components/binding-usage-controller/docs) directory.