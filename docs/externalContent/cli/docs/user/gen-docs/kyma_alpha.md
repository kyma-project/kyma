# kyma alpha

Groups command prototypes for which the API may still change.

## Synopsis

A set of alpha prototypes that may still change. Use them in automation at your own risk.

```bash
kyma alpha <command> [flags]
```

## Available Commands

```text
  app                - Manages applications on the Kubernetes cluster
  hana               - Manages an SAP HANA instance in the Kyma cluster
  kubeconfig         - Manages access to the Kyma cluster
  provision          - Provisions a Kyma cluster on SAP BTP
  reference-instance - Adds an instance reference to a shared service instance
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma](kyma.md)                                                   - A simple set of commands to manage a Kyma cluster
* [kyma alpha app](kyma_alpha_app.md)                               - Manages applications on the Kubernetes cluster
* [kyma alpha hana](kyma_alpha_hana.md)                             - Manages an SAP HANA instance in the Kyma cluster
* [kyma alpha kubeconfig](kyma_alpha_kubeconfig.md)                 - Manages access to the Kyma cluster
* [kyma alpha provision](kyma_alpha_provision.md)                   - Provisions a Kyma cluster on SAP BTP
* [kyma alpha reference-instance](kyma_alpha_reference-instance.md) - Adds an instance reference to a shared service instance
