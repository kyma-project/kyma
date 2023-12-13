# Configuring the Runtime

> **NOTE:** To represent API and Event Definitions of the Applications connected to a Runtime, we recommend that you use the Open Service Broker API.

Runtime Agent periodically requests for the configuration of its Runtime from Compass.
Changes in the configuration for the Runtime are applied by Runtime Agent on the Runtime.

To fetch the Runtime configuration, Runtime Agent calls the [`applicationsForRuntime`](https://github.com/kyma-incubator/compass/blob/master/components/director/pkg/graphql/schema.graphql) query offered by the Compass component called Director.
The response for the query contains Applications assigned for the Runtime.
Each Application contains only credentials that are valid for the Runtime that called the query. Runtime Agent uses the credentials to create Secrets used by Application Gateway. 
Each Runtime Agent can fetch the configurations for Runtimes that belong to its tenant.

Runtime Agent reports back to the Director the Runtime-specific [LabelDefinitions](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-04-labels.md#labeldefinitions), which represent Runtime configuration, together with their values.
Runtime-specific LabelDefinitions are Event Gateway URL and Runtime Console URL.

The data mapping between Director and Kyma looks as follows: 

| **Director (Compass)**    | **Kyma**                      |
|---------------------------|-------------------------------|
| Application               | Application CR                |
| API Bundle                | Service in the Application CR |
| API Definition            | Entry under the service       |