# Configuring the Runtime

Runtime Agent periodically requests for the configuration of Kyma runtime from Unified Customer Landscape (UCL).

To fetch the Kyma runtime configuration, Runtime Agent calls the [`applicationsForRuntime`](https://github.com/kyma-incubator/compass/blob/master/components/director/pkg/graphql/schema.graphql) query offered by the component called the UCL Director.

The response for the query contains Applications (an Application represents an external system) assigned for Kyma runtime.

Each Application contains credentials that are only valid and unique for querying Kyma runtime (another requestor cannot use them). Runtime Agent stores the credentials in Secrets, which are used by Application Gateway to establish a trusted outbound communication to an external system.

This data mapping shows how the retrieved configuration of an Application from the UCL Director is stored in Kyma runtime:

| **UCL Director**    | **Kyma Runtime**                    |
|---------------------------|-------------------------------|
| Application               | Application CR                |
| API Bundle                | Service in the Application CR |
| API Definition            | Entry under the service       |

## Application Name

The name of the Application is used as a key within the Application Connector module and has the following special requirements:

### Uniqueness

The names of Applications assigned to the Runtime must be unique in Kyma Runtime. If they are not unique, the synchronization fails.

### Normalization of Application Names

Runtime Agent can normalize the names of Applications fetched from the UCL Director by converting them to lowercase and removing special characters and spaces.

This feature is controlled by the `isNormalized` label, which can be set on the Runtime in UCL.

When the Runtime is initially labeled with `isNormalized=true`, Runtime Agent normalizes the names of Applications. When the Runtime is initially labeled with `isNormalized=false`, or if the Runtime does not contain such a label, Runtime Agent doesn't normalize the names.

The normalization can lead to non-unique Application names if names are differentiated only by special characters or by different lower or upper case letters.

## Reporting Kyma Runtime Configuration to UCL

Runtime Agent reports back to the Director the Runtime-specific [LabelDefinitions](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-04-labels.md#labeldefinitions), which represent Runtime configuration, together with their values.
Runtime-specific LabelDefinitions are Event Gateway URL and Runtime Console URL.
