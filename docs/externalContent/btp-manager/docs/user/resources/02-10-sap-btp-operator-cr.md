# SAP BTP Operator Custom Resource

The `btpoperators.operator.kyma-project.io` Custom Resource Definition (CRD) is a comprehensive specification that defines the structure and format used to manage the configuration and status of the SAP BTP Operator module within your Kyma environment.

To get the latest CRD in the YAML format, run the following command:

```shell
kubectl get crd btpoperators.operator.kyma-project.io -o yaml
```
You can only have one SAP BTP Operator (BtpOperator) CR. The BtpOperator CR must be in the `kyma-system` namespace, and the resource's name must be 'btpoperator'. Any other BtpOperator CR has the `Warning` state.

## Sample Custom Resource

The following BtpOperator object defines a module:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: BtpOperator
metadata:
  finalizers:
    - operator.kyma-project.io/btp-manager
  labels:
    app.kubernetes.io/created-by: btp-manager
    app.kubernetes.io/instance: btpoperator
    app.kubernetes.io/managed-by: btp-manager
    app.kubernetes.io/name: btpoperator
    app.kubernetes.io/part-of: btp-manager
  name: btpoperator
  namespace: kyma-system
spec: {}
status:
  conditions:
    - lastTransitionTime: '2024-08-08T14:39:01Z'
      message: Module provisioning succeeded
      reason: ReconcileSucceeded
      status: 'True'
      type: Ready
  state: Ready
```

## Custom Resource Parameters

**Spec:** 

You can't configure any entry parameters in the BtpOperator CR.

**Status:**

| No. | CR state             | Condition type       | Condition status     | Condition reason                                            | Remark                                                                                        |
|-----| -------------------- | -------------------- | -------------------- | ----------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| 1   | Ready                | Ready                | true                 | ReconcileSucceeded                                          | Reconciled successfully                                                                       |
| 2   | Ready                | Ready                | true                 | UpdateCheckSucceeded                                        | Update not required                                                                           |
| 3   | Ready                | Ready                | true                 | UpdateDone                                                  | Update done                                                                                   |
| 4   | Processing           | Ready                | false                | ClusterIdChanged                                            | Cluster ID changed                                                                            |
| 5   | Processing           | Ready                | false                | CredentialsNamespaceChanged                                 | Credentials namespace changed                                                                 |
| 6   | Processing           | Ready                | false                | Initialized                                                 | Initial processing or chart is inconsistent                                                   |
| 7   | Processing           | Ready                | false                | Processing                                                  | Final State after deprovisioning                                                              |
| 8   | Processing           | Ready                | false                | UpdateCheck                                                 | Checking for updates                                                                          |
| 9   | Processing           | Ready                | false                | Updated                                                     | Resource has been updated                                                                     |
| 10  | Deleting             | Ready                | false                | HardDeleting                                                | Trying to hard delete                                                                         |
| 11  | Deleting             | Ready                | false                | SoftDeleting                                                | Trying to soft-delete after hard-delete failed                                                |
| 12  | Error                | Ready                | false                | AnnotatingSecretFailed                                      | Annotating the required Secret failed                                                         |
| 13  | Error                | Ready                | false                | ChartInstallFailed                                          | Failure during chart installation                                                             |
| 14  | Error                | Ready                | false                | ChartPathEmpty                                              | No chart path available for processing                                                        |
| 15  | Error                | Ready                | false                | ConsistencyCheckFailed                                      | Failure during consistency check                                                              |
| 16  | Error                | Ready                | false                | DeletionOfOrphanedResourcesFailed                           | Deletion of orphaned resources failed                                                         |
| 17  | Error                | Ready                | false                | GettingConfigMapFailed                                      | Getting ConfigMap failed                                                                      |
| 18  | Error                | Ready                | false                | GettingDefaultCredentialsSecretFailed                       | Getting default credentials Secret failed                                                     |
| 19  | Error                | Ready                | false                | GettingSapBtpServiceOperatorClusterIdSecretFailed           | Getting SAP BTP service operator Cluster ID Secret failed                                     |
| 20  | Error                | Ready                | false                | GettingSapBtpServiceOperatorConfigMapFailed                 | Getting SAP BTP service operator ConfigMap failed                                             |
| 21  | Error                | Ready                | false                | InconsistentChart                                           | Chart is inconsistent, reconciliation initialized                                             |
| 22  | Error                | Ready                | false                | InvalidSecret                                               | `sap-btp-manager` Secret does not contain required data - create proper Secret                |
| 23  | Error                | Ready                | false                | PreparingInstallInfoFailed                                  | Error while preparing installation information                                                |
| 24  | Error                | Ready                | false                | ProvisioningFailed                                          | Provisioning failed                                                                           |
| 25  | Error                | Ready                | false                | ReconcileFailed                                             | Reconciliation failed                                                                         |
| 26  | Error                | Ready                | false                | ResourceRemovalFailed                                       | Some resources can still be present due to errors while deprovisioning                        |
| 27  | Error                | Ready                | false                | StoringChartDetailsFailed                                   | Failure of storing chart details                                                              |
| 28  | Warning              | Ready                | false                | MissingSecret                                               | `sap-btp-manager` Secret was not found - create proper Secret                                 |
| 29  | Warning              | Ready                | false                | ServiceInstancesAndBindingsNotCleaned                       | Deprovisioning blocked because of ServiceInstances and/or ServiceBindings existence           |
| 30  | Warning              | Ready                | false                | WrongNamespaceOrName                                        | Wrong namespace or name                                                                       |

