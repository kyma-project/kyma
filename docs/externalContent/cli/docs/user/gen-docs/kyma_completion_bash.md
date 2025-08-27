# kyma completion bash

Generate the autocompletion script for bash.

## Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(kyma completion bash)

To load completions for every new session, execute once:

#### Linux:

	kyma completion bash > /etc/bash_completion.d/kyma

#### macOS:

	kyma completion bash > $(brew --prefix)/etc/bash_completion.d/kyma

You will need to start a new shell for this setup to take effect.


```bash
kyma completion bash
```

## Flags

```text
      --no-descriptions         disable completion descriptions
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma completion](kyma_completion.md) - Generate the autocompletion script for the specified shell
