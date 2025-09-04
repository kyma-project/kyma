# kyma alpha reference-instance

Adds an instance reference to a shared service instance.

## Synopsis

Use this command to add an instance reference to a shared service instance in the Kyma cluster.

```bash
kyma alpha reference-instance [flags]
```

## Flags

```text
      --btp-secret-name string       name of the BTP secret containing credentials to another subaccount Service Manager:
                                     https://github.com/SAP/sap-btp-service-operator/blob/main/README.md#working-with-multiple-subaccounts
      --instance-id string           ID of the instance
      --label-selector stringSlice   Label selector for filtering instances (default "[]")
      --name-selector string         Instance name selector for filtering instances
      --namespace string             Namespace of the reference instance (default "default")
      --offering-name string         Offering name
      --plan-selector string         Plan name selector for filtering instances
      --reference-name string        Name of the reference
  -h, --help                         Help for the command
      --kubeconfig string            Path to the Kyma kubeconfig file
      --show-extensions-error        Prints a possible error when fetching extensions fails
      --skip-extensions              Skip fetching extensions from the cluster
```

## See also

* [kyma alpha](kyma_alpha.md) - Groups command prototypes for which the API may still change
