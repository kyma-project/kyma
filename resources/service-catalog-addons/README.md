# Service Catalog Add-ons

## Overview

The Service Catalog Add-ons provide Kyma add-ons to the [Service Catalog](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/README.md).

These add-ons consist of the following items:
* Kyma Console views related to the Service Catalog
* Binding Usage Controller which extends the Service Catalog with the additional logic

### Views

The Service Catalog Add-ons provide the following views to the Kyma Console:

* Catalog UI
* Instances UI
* Brokers UI

### Service Binding Usage Controller

The Service Binding Usage Controller allows you to inject Secrets to a given application. For this purpose, it introduces two custom resources (CRs):

* [ServiceBindingUsage](../../docs/service-catalog/06-01-service-binding-usage.md)
* [UsageKind](../../docs/service-catalog/06-02-usage-kind.md)

The Service Binding Usage Controller chart provides two default UsageKinds to Kyma:

* [Deployment](charts/service-binding-usage-controller/templates/deployment-usage-kind.yaml)

For more detailed information, go to the [Service Binding Usage Controller](https://github.com/kyma-project/kyma/tree/master/components/service-binding-usage-controller/docs) directory.
