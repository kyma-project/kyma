---
title: Kyma CLI
type: UI
---

Kyma CLI is a command-line tool that supports Kyma developers. It provides a set of commands and flags you can use to: 

- Provision a cluster locally or on cloud providers, such as GCP or Azure, or use Gardener to set up and easily manage your clusters.
- Install, manage, and test Kyma.

Kyma CLI is always released in parallel with Kyma to support the latest features, which also means that older Kyma versions can no longer be supported by the Kyma CLI.
This Kyma CLI version is compatible with the corresponding Kyma release and the previous release, but it's incompatible with versions before the previous ones.

Kyma CLI comes with a set of commands, each of which has its own specific set of flags. Use them to provision the cluster locally or using a chosen cloud provider, install, and test Kyma.

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

See [the full list of commands and flags](/cli/commands/).