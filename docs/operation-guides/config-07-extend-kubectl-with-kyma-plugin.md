---
title: Use Kyma CLI as kubectl plugin
type: Configuration
---

You can extend kubectl to support Kyma CLI and its commands. 

## Prerequisites

You are using Kubernetes version 1.12.0 or higher.

## Steps

To use Kyma CLI as the [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), perform the following steps:

1. Rename the `kyma` binary to `kubectl-kyma` and place it anywhere in your **{PATH}**:

    ```bash
    sudo mv ./kyma /usr/local/bin/kubectl-kyma
    ```

2. Run `kubectl plugin list` command to see your plugin on the list of all available plugins.

3. Invoke your plugin as a kubectl command:

    ```bash
    $ kubectl kyma version
    Kyma CLI version: v0.6.1
    Kyma cluster version: 1.0.0
    ```
