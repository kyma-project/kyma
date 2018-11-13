<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-project/kyma/master/logo.png" width="235">
</p>

[![Go Report Card](https://goreportcard.com/badge/kyma-project/kyma)](https://goreportcard.com/report/github.com/kyma-project/kyma)

## Overview

Kyma allows you to connect applications and third-party services in a cloud-native environment. Use it to create extensions for the existing systems, regardless of the language they are written in. Customize extensions with minimum effort and time devoted to learning their configuration details.

Go to the [Kyma project website](https://kyma-project.io/) to learn more about the product, its features, and components.

## Installation

Install Kyma [locally](https://kyma-project.io/docs/root/kyma#getting-started-local-kyma-installation) and on a [cluster](https://kyma-project.io/docs/root/kyma#getting-started-cluster-kyma-installation).

## Usage

Kyma comes with the ready-to-use code snippets that you can use to test the extensions and the core functionality. See the list of existing examples in the [`examples`](https://github.com/kyma-project/examples) repository.

## Development

Develop on your remote repository forked from the original repository in Go.
See the example that uses the [`ui-api-layer`](./components/ui-api-layer) project located in the `components` directory but applies to any Go project. This set of instructions uses the recommended [`git workflow`](https://github.com/kyma-project/community/blob/master/git-workflow.md) and the general [contribution flow](https://github.com/kyma-project/community/blob/master/CONTRIBUTING.md#contribute-code-or-content). Read also the [`CONTRIBUTING.md`](CONTRIBUTING.md) document that includes the contributing rules specific for this repository.

Follow these steps:

> **NOTE:** The example assumes you have the `$GOPATH` already set.

1. Fork the repository in GitHub.

2. Clone the fork to your `$GOPATH` workspace. Use this command to create the folder structure and clone the repository under the correct location:

```
git clone git@github.com:{GitHubUsername}/kyma.git $GOPATH/src/github.com/kyma-project/kyma
```

Follow the steps described in the [`git-workflow.md`](https://github.com/kyma-project/community/blob/master/git-workflow.md#steps) document to configure your fork.

3. Install dependencies.

Go to the main directory of the project in your workspace location and install the required dependencies:

```
$ cd components/ui-api-layer
$ dep ensure -vendor-only
```

4. Build the project.

Every project runs differently. Follow instructions in the main `README.md` document of the given project to build it.

5. Create a branch and start to develop.

Do not forget about creating unit and acceptance tests if needed. For the unit tests, follow the instructions specified in the main `README.md` document of the given project. For the details concerning the acceptance tests, go to the corresponding directory inside the `tests` directory.
Find the information on how to run changes on the cluster without a Docker image in the [Develop a service locally without using Docker](https://kyma-project.io/docs/latest/root/kyma#getting-started-develop-a-service-locally-without-using-docker) document.

>**NOTE:** For more details about testing, go to the [Testing Kyma](https://kyma-project.io/docs/latest/root/kyma#details-testing-kyma) document.

6. Test your changes.

### Project structure

The repository has the following structure:

```
  ├── .github                     # Pull request and issue templates             
  ├── components                  # Source code of all Kyma components                                                
  ├── docs                        # Documentation source files
  ├── installation                # Installation scripts     
  ├── resources                   # Helm charts and Kubernetes resources for the Kyma installation
  ├── tests                       # Acceptance tests
  └── tools                       # Source code of utilities used, for example, for the installation and testing
  ```
