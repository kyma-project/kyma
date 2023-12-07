---
title: What are the UIs available in Kyma?
---

Kyma provides two interfaces that you can use for interactions:

- **Kyma Dashboard** - a web-based administrative UI that you can use to manage the basic Kubernetes resources.
- **Kyma CLI** - a CLI to execute various Kyma tasks, such as installing or upgrading Kyma.

# Kyma Dashboard

## Purpose

Kyma uses [Busola](https://github.com/kyma-project/busola) as a central administration dashboard, which provides a graphical overview of your cluster and deployments.

You can deploy microservices, create Functions, and manage their configurations. You can also use it to register cloud providers for additional services, create instances of these services, and use them in your microservices or Functions.

## Integration

Busola is a web-based UI for managing resources within Kyma or any Kubernetes cluster. Busola has a dedicated Node.js backend, which is a proxy for a [Kubernetes API server](https://kubernetes.io/docs/concepts/overview/components/#kube-apiserver).

# Kyma CLI

## Purpose

Kyma CLI is a command-line tool that supports Kyma developers. It provides a set of commands and flags you can use to:

- Provision a cluster locally or on cloud providers, such as GCP or Azure, or use Gardener to set up and easily manage your clusters.
- Install, manage, and test Kyma.
- Manage your Functions.

## Compatibility

Kyma CLI is always released in parallel with Kyma to support the latest features, which also affects backwards compatibility. The current Kyma CLI version supports the corresponding Kyma release and the previous release, but it's incompatible with Kyma versions before the previous ones.

## Commands and Flags

Kyma CLI comes with a set of commands, each of which has its own specific set of flags. Use them to provision the cluster locally or using a chosen cloud provider, install, and test Kyma.

See [the full list of commands and flags](https://github.com/kyma-project/cli/tree/main/docs/gen-docs).

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
