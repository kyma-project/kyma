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

### Binding Usage Controller

The Binding Usage Controller allows you to inject Secrets to a given application. For this purpose, it introduces two custom resources (CRs):

* [ServiceBindingUsage](../../../../docs/service-catalog/docs/06-01-service-binding-usage.md)
* [UsageKind](../../../../docs/service-catalog/docs/06-02-usage-kind.md)

The Binding Usage Controller chart provides two default UsageKinds to Kyma:

* [Function](charts/binding-usage-controller/templates/function-usage-kind.yaml)
* [Deployment](charts/binding-usage-controller/templates/deployment-usage-kind.yaml)

For more detailed information, go to the [Binding Usage Controller](https://github.com/kyma-project/kyma/tree/master/components/binding-usage-controller/docs) directory.
