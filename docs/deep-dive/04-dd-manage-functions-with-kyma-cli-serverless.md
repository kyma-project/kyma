---
title: Manage Functions with Kyma CLI
type: Deep Dive
---

This tutorial shows how to use the available CLI commands to manage Functions in Kyma. You will see how to:

1. Create local files that contain the basic configuration for a sample "Hello World" Python Function (`kyma init function`).
2. Generate a Function custom resource (CR) from these files and apply it on your cluster (`kyma apply function`).
3. Fetch the current state of your Function's cluster configuration after it was modified (`kyma sync function`).

This tutorial is based on a sample Python Function run on a lightweight [k3d](https://k3d.io/) cluster.

## Prerequisites

Before you start, make sure you have the following tools installed:

- [Docker](https://www.docker.com/)
- [Kyma CLI](#installation-install-kyma-cli)
- Kyma installed [locally](https://kyma-project.io/docs/main/root/kyma/#installation-install-kyma-locally) or [on a cluster](https://kyma-project.io/docs/main/root/kyma/#installation-install-kyma-on-a-cluster)


## Steps

Follow these steps:

1. Run the `init` Kyma CLI command to create local files with the default configuration for a Python Function. Go to the folder in which you want to initiate the workspace content and run this command:

  ```bash
  kyma init function --runtime python38 --name {FUNCTION_NAME}
  ```

  Alternatively, use the `--dir {FULL_FOLDER_PATH}` flag to point to the directory where you want to create the Function's source files.

  > **NOTE:** Python 3.8 is only one of the available runtimes. Read about all [supported runtimes and sample Functions to run on them](https://kyma-project.io/docs/main/components/serverless/#details-runtimes).

  The `init` command creates the following files in your workspace folder:

  - `config.yaml`	with the Function's configuration

> **NOTE:** See the [detailed description](#details-function-configuration-file) of all fields available in the Function's configuration file.

  - `handler.py` with the Function's code and the simple "Hello World" logic
  - `requirements.txt` with an empty file for your Function's custom dependencies

  This command also sets **sourcePath** in the `config.yaml` file to the full path of the workspace folder:

  ```yaml
  name: my-function
  namespace: default
  runtime: python38
  source:
      sourceType: inline
      sourcePath: {FULL_PATH_TO_WORKSPACE_FOLDER}
  ```

2. Run the `apply` Kyma CLI command to create a Function CR in the YAML format on your cluster:

  ```bash
  kyma apply function
  ```

  > **TIP:** To apply a Function from a different location, use the `--filename` flag followed by the full path to the `config.yaml` file.

  Alternatively, use the `--dry-run` flag to list the file that will be created before you apply it. You can also preview the file's content in the format of your choice by adding the `--output {FILE_FORMAT}` flag, such as `--output yaml`.

3. Once applied, view the Function's details on the cluster:

  ```bash
  kubectl describe function {FUNCTION_NAME}
  ```

4. Change the Function's source code on the cluster to return "Hello Serverless!":

  a) Edit the Function:

  ```bash
  kubectl edit function {FUNCTION_NAME}
  ```

  b) Modify **source** as follows:

  ```yaml
  ...
  spec:
    runtime: python38
    source: |-
      def main(event, context):
          return "Hello Serverless!"
  ```

5. Fetch the content of the resource to synchronize your local workspace sources with the cluster changes:

  ```bash
  kyma sync function {FUNCTION_NAME}
  ```

6. Check the local `handler.py` file with the Function's code to make sure that the cluster changes were fetched:

  ```bash
  cat handler.py
  ```

  This command returns the result confirming that the local sources were synchronized with cluster changes:

  ```py
  def main(event, context):
      return "Hello Serverless!"
  ```
