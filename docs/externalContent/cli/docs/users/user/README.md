# Kyma CLI

Kyma CLI is an essential tool for application developers who want to get started quickly and efficiently with SAP BTP, Kyma runtime. Designed to streamline workflows, it simplifies complex tasks, enabling developers to deploy and manage applications easily.

> [!WARNING]
> The alpha group commands are still in development, which means their functions and API may be modified over time. We encourage you to explore them, but keep in mind that changes may occur.

## What is Kyma CLI?

With a single command, you can push a simple application to a Kyma cluster. Kyma CLI builds the container image from source code or Dockerfile, pushes it to the in-cluster registry, and applies the necessary Kubernetes resources, eliminating the manual setup process and accelerating your development.

Kyma CLI provides a hands-on experience when interacting with Kyma’s extensions to Kubernetes API, making it easier to explore, test, and fine-tune Kyma's custom resources directly from the command line.

Kyma CLI also provides a set of commands to manage Kyma modules efficiently. You can manage, deploy, and configure modules seamlessly. With the Kyma CLI module commands, you can list available and installed modules, and add or delete them. Modules can be deployed with the default or custom configuration.

## Features

The Kyma CLI provides the following features:

- Simplified module management.
- Automated deployments.
- Simplified SAP BTP service instance bindings.
- Commands providing useful automation.

## Install Kyma CLI

### Stable Release

To get the latest stable Kyma CLI for MacOS or Linux, run the following script from the command line:

<!-- tabs:start -->

#### **Homebrew**

```sh
brew install kyma-cli
```

#### **GitHub releases**

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_stable.sh" | sh -
kyma
```

<!-- tabs:end -->

### Latest Build

Download the latest (stable or unstable pre-release) v3 build from the [releases](https://github.com/kyma-project/cli/releases) assets.

To get the latest Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_latest.sh" | sh -
kyma
```

### Nightly Build

Download the latest build from the main branch from [0.0.0-dev](https://github.com/kyma-project/cli/releases/tag/0.0.0-dev) release assets.

To get Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_nightly.sh" | sh -
kyma
```

> [!TIP]
> Before you use Kyma CLI, we strongly recommend adding the autocomplete formula to your shell (`bash`, `fish`, `PowerShell`, or `zsh`).
> Run the following command to import the Kyma CLI autocomplete formula for the `zsh` shell and enable autocomplete with the tab key hit:
> `source <(kyma completion zsh)`

## Related Information

- [Kyma CLI tutorials](tutorials/README.md)
