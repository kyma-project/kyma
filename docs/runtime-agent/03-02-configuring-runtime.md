---
title: Configuring the Runtime
type: Details
---

> **NOTE:** To represent API and Event Definitions of the connected Applications on Runtime, Open Service Broker API usage is recommended.

In Kyma Runtime, during Runtime configuration, Application's Packages are integrated into the [Service Catalog](components/service-catalog) using [Application](components/application-connector#custom-resource-application) custom resources and [Application Broker](components/application-connector#architecture-application-broker). By default, a single Application is represented as a [ServiceClass](components/service-catalog/#architecture-resources), and a single Package is represented as a [ServicePlan](components/service-catalog/#architecture-resources) in the Service Catalog. Read more about [API Packages](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-packages-api.md) document.

Runtime Agent periodically requests for the configuration of its Runtime from Compass. Changes in the configuration for the Runtime are applied by the Runtime Agent on the Runtime.

To fetch the Runtime configuration, Runtime Agent calls the [`applicationsForRuntime`](https://github.com/kyma-incubator/compass/blob/master/components/director/pkg/graphql/schema.graphql) query offered by the Compass component called Director. The response for the query contains a page with the list of Applications assigned for the Runtime and info about the next page. Each Application will contain only credentials that are valid for the Runtime that called the query. Each Runtime Agent can fetch the configurations for Runtimes that belong to its tenant, there is no validation if the Runtime Agent is fetching the configuration for the Runtime on which it runs.

Runtime Agent reports back to the Director the Runtime-specific [LabelDefinitions](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-02-labels.md#labeldefinitions) that represent Runtime configuration together with their values.

Runtime-specific LabelDefinitions:

- Events Gateway URL
- Runtime Console URL