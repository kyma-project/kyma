# Manage Functions with Kyma CLI

This tutorial shows how to use the available CLI commands to manage Functions in Kyma. You will see how to:

1. Generate local files that contain source code and dependencies for a sample "Hello World" Python Function (`kyma alpha function init`).
2. Create a Function custom resource (CR) from these files and apply it on your cluster (`kyma alpha function create`).
3. List all Function CRs from the cluster (`kyma alpha function get`).
4. Delete previously created Function CR from the cluster (`kyma alpha function delete`).

This tutorial is based on a sample Python Function run in a lightweight [k3d](https://k3d.io/) cluster.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Docker](https://www.docker.com/)
- [Kyma CLI](https://github.com/kyma-project/cli)
- Serverless module

## Procedure

Follow these steps:

1. To create local files with the default configuration for a Python Function, go to the folder in which you want to initiate the workspace content and run the `init` Kyma CLI command:

  ```bash
  kyma alpha function init --runtime python312
  ```

  You can also use the `--dir {FULL_FOLDER_PATH}` flag to point to the directory where you want to create the Function's source files.

  > [!NOTE]
  > Python 3.12 is only one of the available runtimes. Read about all [supported runtimes and sample Functions to run on them](../technical-reference/07-10-sample-functions.md).

  The `init` command creates these files in your workspace folder:

- `handler.py` - the Function's code and the simple "Hello World" logic
- `requirements.txt` - an empty file for your Function's custom dependencies

  After your files are successfully created, you will see the following output message with suggestions on how you can run the code:

  ```text
  Next steps:
  * update output files in your favorite IDE
  * create Function, for example:
    kyma alpha function create python312 --runtime python312 --source handler.py --dependencies requirements.txt
  ```

  > [!NOTE]
  > Now is a good time to customize your Python runtimes source code. You can do it by editing generated files in your favorite editor.

2. Run the `create` Kyma CLI command to create a Function CR in the YAML format in your cluster based on the suggestion from the previous command:

  ```bash
  kyma alpha function create python312 --runtime python312 --source handler.py --dependencies requirements.txt
  ```

3. Once applied, check the Function's state in the cluster:

  ```bash
  kyma alpha function get
  ```

4. Delete Function from the cluster:

  ```bash
  kyma alpha function delete {FUNCTION_NAME}
  ```
