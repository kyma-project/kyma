---
title: Configuring the Runtime
type: Details
---

> **NOTE:** To represent API and Event Definitions of the Applications connected to a Runtime, Open Service Broker API usage is recommended.

In a Kyma Runtime, during Runtime configuration, Application's Bundles are integrated into [Service Catalog](components/service-catalog) using [Application](components/application-connector#custom-resource-application) custom resources and [Application Broker](components/application-connector#architecture-application-broker).
By default, a single Application is represented as a [ServiceClass](components/service-catalog/#architecture-resources), and a single Bundle is represented as a [ServicePlan](components/service-catalog/#architecture-resources) in Service Catalog.
Refer to the documentation to learn more about [API Bundles](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-bundles-api.md).

Runtime Agent periodically requests for the configuration of its Runtime from Compass.
Changes in the configuration for the Runtime are applied by the Runtime Agent on the Runtime.

To fetch the Runtime configuration, Runtime Agent calls the [`applicationsForRuntime`](https://github.com/kyma-incubator/compass/blob/master/components/director/pkg/graphql/schema.graphql) query offered by the Compass component called Director.
The response for the query contains Applications assigned for the Runtime.
Each Application contains only credentials that are valid for the Runtime that called the query.
Each Runtime Agent can fetch the configurations for Runtimes that belong to its tenant.

Runtime Agent reports back to the Director the Runtime-specific [LabelDefinitions](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-02-labels.md#labeldefinitions), which represent Runtime configuration, together with their values.
Runtime-specific LabelDefinitions are Events Gateway URL and Runtime Console URL.