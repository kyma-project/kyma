# kyma completion zsh

Generate the autocompletion script for zsh.

## Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(kyma completion zsh)

To load completions for every new session, execute once:

#### Linux:

	kyma completion zsh > "${fpath[1]}/_kyma"

#### macOS:

	kyma completion zsh > $(brew --prefix)/share/zsh/site-functions/_kyma

You will need to start a new shell for this setup to take effect.


```bash
kyma completion zsh [flags]
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
