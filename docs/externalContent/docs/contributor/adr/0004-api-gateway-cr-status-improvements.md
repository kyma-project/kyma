# API Gateway Custom Resource Status Improvements

## Status
Accepted

## Context
The API Gateway CR status doesn't contain enough information to easily debug without directly accessing Kyma API Gateway Operator's logs. We would like to make it easier both for the user and the Kyma API Gateway team to view the state of the API Gateway module.

## Decision

For the sake of simplicity, we implement the condition of type `Ready` as the main condition that supports multiple reasons. This way, we can reduce the number of condition types. If we use the [SetStatusCondition function](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/meta#SetStatusCondition) of [k8s.io/apimachinery](https://github.com/kubernetes/apimachinery), the **lastTransitionTime** is only updated when the status of the condition changes.

To support multiple reasons for the status `False`, we decided to set the status of the `Ready` condition to `Unknown` at the beginning of the reconciliation. This allows us to force the **lastTransitionTime** of the `Ready` condition to be always updated.

In addition, the `Ready` condition reflects one reconciliation status. Therefore, it is a good idea to set it to `Unknown` at the beginning.

For `KymaGatewayDeletionBlocked` and `DeletionBlockedExistingResources` reasons of the `Ready` condition with the status `Warning` we report up to 5 custom resources that are blocking the deletion. The listing format for each is `<kind>/<name>`.

Conditions:

| CR state   | Type                         | Status | Reason                                | Message                                                                                        |
|------------|------------------------------|--------|---------------------------------------|------------------------------------------------------------------------------------------------|
| Ready      | Ready                        | True   | ReconcileSucceeded                    | Reconciliation succeeded                                                                       |
| Error      | Ready                        | False  | ReconcileFailed                       | Reconciliation failed                                                                          |
| Error      | Ready                        | False  | OlderCRExists                         | API Gateway CR is not the oldest one and does not represent the module state                   |
| Error      | Ready                        | False  | CustomResourceMisconfigured 			 | API Gateway CR has invalid configuration                                                       |
| Error      | Ready                        | False  | DependenciesMissing                   | Module dependencies missing                                                                   |
| Processing | Ready                        | False  | KymaGatewayReconcileSucceeded         | Kyma Gateway reconciliation succeeded                                                          |
| Error      | Ready                        | False  | KymaGatewayReconcileFailed            | Kyma Gateway reconciliation failed                                                             |
| Warning    | Ready                        | False  | KymaGatewayDeletionBlocked            | Kyma Gateway deletion blocked because of the existing custom resources: ...                        |
| Processing | Ready                        | False  | OathkeeperReconcileSucceeded          | Ory Oathkeeper reconciliation succeeded                                                        |
| Error      | Ready                        | False  | OathkeeperReconcileFailed             | Ory Oathkeeper reconciliation failed                                                           |
| Warning    | Ready                        | False  | DeletionBlockedExistingResources      | API Gateway deletion blocked because of the existing custom resources: ...                         |

## Consequences
This architectural decision allows our team and customers to monitor the state of the API Gateway module on the cluster more easily. Previously, accessing the API Gateway module logs was often required to gain better visibility into the issues occurring in the cluster.

Since it becomes a part of our API, we must ensure the stability.