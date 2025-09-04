# kyma alpha hana

Manages an SAP HANA instance in the Kyma cluster.

## Synopsis

Use this command to manage an SAP HANA instance in the Kyma cluster.

```bash
kyma alpha hana <command> [flags]
```

## Available Commands

```text
  map - Maps an SAP HANA instance to the Kyma cluster
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma alpha](kyma_alpha.md)                   - Groups command prototypes for which the API may still change
* [kyma alpha hana map](kyma_alpha_hana_map.md) - Maps an SAP HANA instance to the Kyma cluster
