---
title: Install Kyma CLI
type: Tasks
---

<!-- Installing the Kyma CLI is a "basic task"; users need the CLI to get started -->

You can easily install Kyma CLI on macOS, Linux, or Windows. To do so, perform the instructions described in the following sections.

## macOS

To install Kyma CLI on macOS, run:

```bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Darwin_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && sudo mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz
```

### Homebrew

To install Kyma CLI using Homebrew, run:

```bash
brew install kyma-cli
```

If the Homebrew team does not update the Homebrew formula of the CLI within three days of the release, follow this [guide](https://github.com/Homebrew/brew/blob/master/docs/How-To-Open-a-Homebrew-Pull-Request.md) to update it manually to the most recent version. For a sample Homebrew Kyma CLI formula version bump, see [this PR](https://github.com/Homebrew/homebrew-core/pull/52375).

## Linux

To install Kyma CLI on Linux, run:

```bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Linux_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && sudo mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz
```

## Windows

To install Kyma CLI on Windows, download and unzip the [release artifact](https://github.com/kyma-project/cli/releases). Remember to adjust your **PATH** environment variable.

```PowerShell
${KYMA_VERSION}=1.2.0
Invoke-WebRequest -OutFile kyma.zip https://github.com/kyma-project/cli/releases/download/${KYMA_VERSION}/kyma_Windows_x86_64.zip

Expand-Archive -Path kyma.zip -DestinationPath .\kyma-cli

cd kyma-cli
```

### Chocolatey (Windows)

To install Kyma CLI on Windows using [Chocolatey](https://www.chocolatey.org), run:

```PowerShell
choco install kyma-cli
```

You don't have to bump Kyma CLI Chocolatey package manually with each new release, as it includes a script that automatically checks for new releases and updates the package to the latest one.

Still, the package requires some maintenance to keep its dedicated [site](https://chocolatey.org/packages/kyma-cli) at`chocolatey.org` up to date. This means you should regularly update the description, details, screenshots, etc. To keep the site up to date, submit a pull request to [Chocolatey's GitHub repository](https://github.com/dgalbraith/chocolatey-packages/tree/master/automatic/kyma-cli).

## Other

To install a different release version, change the path to point to the desired version and architecture:

```bash
curl -Lo kyma.tar.gz https://github.com/kyma-project/cli/releases/download/${KYMA_VERSION}/kyma_${ARCH}.tar.gz
```
