# kyma completion

Generate the autocompletion script for the specified shell.

## Synopsis

Generate the autocompletion script for kyma for the specified shell.
See each sub-command's help for details on how to use the generated script.


```bash
kyma completion
```

## Available Commands

```text
  bash       - Generate the autocompletion script for bash
  fish       - Generate the autocompletion script for fish
  powershell - Generate the autocompletion script for powershell
  zsh        - Generate the autocompletion script for zsh
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma](kyma.md)                                             - A simple set of commands to manage a Kyma cluster
* [kyma completion bash](kyma_completion_bash.md)             - Generate the autocompletion script for bash
* [kyma completion fish](kyma_completion_fish.md)             - Generate the autocompletion script for fish
* [kyma completion powershell](kyma_completion_powershell.md) - Generate the autocompletion script for powershell
* [kyma completion zsh](kyma_completion_zsh.md)               - Generate the autocompletion script for zsh
