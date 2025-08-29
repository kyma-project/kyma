# APIGateway Custom Resource

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data that APIGateway Controller uses to configure the API Gateway resources. Applying the custom resource (CR) triggers the installation of API Gateway resources, and deleting it triggers the uninstallation of those resources. The default CR has the name `default`.

To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apigateways.operator.kyma-project.io -o yaml
```

You are only allowed to have one APIGateway CR. If there are multiple APIGateway CRs in the cluster, the oldest one reconciles the module. Any additional APIGateway CR is placed in the `Warning` state.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Field                 | Required | Description                                                                                                                                    |
|-----------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| **enableKymaGateway** | **NO**   | Specifies whether the default [Kyma Gateway](./04-10-kyma-gateway.md), named `kyma-gateway`, should be created in the `kyma-system` namespace. |

**Status:**

| Parameter                         | Type       | Description                                                                                                                        |
|-----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                     |

## APIGateway CR's State

|     Code     | Description                                                                                                                                                                                                         |
|:------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|   `Ready`    | APIGateway Controller finished reconciliation.                                                                                                                                                                                 |
| `Processing` | APIGateway Controller is reconciling resources.                                                                                                                                                                                |
|  `Deleting`  | APIGateway Controller is deleting resources.                                                                                                                                                                                   |
|   `Error`    | An error occurred during the reconciliation. The error is rather related to the API Gateway module than the configuration of your resources.                                                                        |
|  `Warning`   | An issue occurred during the reconciliation that requires your attention. Check the status.description message to identify the issue and make the necessary corrections to the APIGateway CR or any related resources. |

## APIGateway CR's Status Conditions

| CR state   | Type  | Status  | Reason                           | Message                                                                      |
|------------|-------|---------|----------------------------------|------------------------------------------------------------------------------|
| `Ready`      | `Ready` | `Unknown` | `ReconcileProcessing`              | Reconciliation processing.                                                    |
| `Ready`      | `Ready` | `True`    | `ReconcileSucceeded`               | Reconciliation succeeded.                                                     |
| `Error`      | `Ready` | `False`   | `ReconcileFailed`                  | Reconciliation failed.                                                        |
| `Error`      | `Ready` | `False`   | `OlderCRExists`                    | APIGateway CR is not the oldest one and does not represent the module state. |
| `Error`      | `Ready` | `False`   | `CustomResourceMisconfigured`      | APIGateway CR has invalid configuration.                                     |
| `Error`      | `Ready` | `False`   | `DependenciesMissing`              | Module dependencies missing.                                                  |
| `Processing` | `Ready` | `False`   | `KymaGatewayReconcileSucceeded`    | Kyma Gateway reconciliation succeeded.                                        |
| `Error`      | `Ready` | `False`   | `KymaGatewayReconcileFailed`       | Kyma Gateway reconciliation failed.                                           |
| `Warning`    | `Ready` | `False`   | `KymaGatewayDeletionBlocked`       | Kyma Gateway deletion blocked because of the existing custom resources: ...  |
| `Processing` | `Ready` | `False`   | `OathkeeperReconcileSucceeded`     | Ory Oathkeeper reconciliation succeeded.                                      |
| `Error`      | `Ready` | `False`   | `OathkeeperReconcileFailed`        | Ory Oathkeeper reconciliation failed.                                         |
| `Warning`    | `Ready` | `False`   | `DeletionBlockedExistingResources` | API Gateway deletion blocked because of the existing custom resources: ...   |