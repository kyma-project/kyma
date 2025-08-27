# kyma alpha provision

Provisions a Kyma cluster on SAP BTP.

## Synopsis

Use this command to provision a Kyma environment on SAP BTP.

```bash
kyma alpha provision [flags]
```

## Flags

```text
      --cluster-name string       Name of the Kyma cluster (default "kyma")
      --credentials-path string   Path to the CIS credentials file
      --environment-name string   Name of the SAP BTP environment (default "kyma")
      --owner string              Email of the Kyma cluster owner
      --parameters string         Path to the JSON file with Kyma configuration
      --plan string               Name of the Kyma environment plan, e.g trial, azure, aws, gcp (default "trial")
      --region string             Name of the region of the Kyma cluster
  -h, --help                      Help for the command
      --kubeconfig string         Path to the Kyma kubeconfig file
      --show-extensions-error     Prints a possible error when fetching extensions fails
      --skip-extensions           Skip fetching extensions from the cluster
```

## See also

* [kyma alpha](kyma_alpha.md) - Groups command prototypes for which the API may still change
