# Kyma CLI

> [!WARNING]
> The Kyma CLI version `v2`, with all commands available within this version, is deprecated. We've started designing the `v3` commands that will be first released within the `alpha` command group.
> Read more about the decision [here](https://github.com/kyma-project/community/issues/872).

## Purpose

Kyma CLI is a command-line tool that supports Kyma developers. It provides a set of commands and flags you can use to:

- Provision a cluster locally or on cloud providers, such as GCP or Azure, or use Gardener to set up and easily manage your clusters.
- Install, manage, and test Kyma.
- Manage your Functions.

## Compatibility

Kyma CLI is always released in parallel with Kyma to support the latest features, which also affects backwards compatibility. The current Kyma CLI version supports the corresponding Kyma release and the previous release, but it's incompatible with Kyma versions before the previous ones.

## Commands and Flags

Kyma CLI comes with a set of commands, each of which has its own specific set of flags. Use them to provision the cluster locally or using a chosen cloud provider, install, and test Kyma.

See [the full list of commands and flags](https://github.com/kyma-project/cli/tree/release-2.20/docs/gen-docs).

## Syntax

For the commands and flags to work, they must follow this syntax:

```bash
kyma {COMMAND} {FLAGS}
```

- **{COMMAND}** specifies the operation you want to perform, such as provisioning the cluster or deploying Kyma.
- **{FLAGS}** specifies optional flags you can use to enrich your command.

See the example:

```bash
kyma deploy -s main
```
