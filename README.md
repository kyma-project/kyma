<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-project/kyma/master/logo.png" width="235">
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/kyma)](https://goreportcard.com/report/github.com/kyma-project/kyma)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2168/badge)](https://bestpractices.coreinfrastructure.org/projects/2168)
[![Slack](https://img.shields.io/badge/slack-@kyma--community-yellow.svg)](http://slack.kyma-project.io)
[![Twitter](https://img.shields.io/badge/twitter-@kymaproject-blue.svg)](https://twitter.com/kymaproject)

## Overview

**Kyma** `/kee-ma/` is a platform for extending applications with microservices and [serverless](https://kyma-project.io/docs/components/serverless/#overview-overview) functions. It provides CLI and UI through which you can connect your application to a Kubernetes cluster and expose the application's API or events securely thanks to the in-built [Application Connector](https://kyma-project.io/docs/components/application-connector/#overview-overview). You can then implement the business logic you require by creating microservices or serverless functions, and triggering them to react to particular events or calls to your application's API. To limit the time spent on coding, use the in-built cloud services from the [Service Catalog](https://kyma-project.io/docs/components/service-catalog/#overview-overview), exposed by [Service Brokers](https://kyma-project.io/docs/components/service-catalog/#service-brokers-service-brokers) from such cloud providers as GCP, Azure, and AWS.

<p align="center">
<a href="https://youtu.be/kP7mSELIxXw">
<img src="./docs/kyma/assets/withoutprov4.gif" style="max-width:100%;">
</a>
</p>

Go to the [Kyma project website](https://kyma-project.io/) to learn more about our project, its features, and components.

## Installation

Install Kyma locally or on a cluster. See the [Installation guides](https://kyma-project.io/docs/root/kyma#installation-installation) for details.

## Usage

Kyma comes with the ready-to-use code snippets that you can use to test the extensions and the core functionality. See the list of existing examples in the [`examples`](https://github.com/kyma-project/examples) repository.

## Development

Develop on your remote repository forked from the original repository in Go.
See the example that uses the [`console-backend-service`](components/console-backend-service) project located in the `components` directory but applies to any Go project. This set of instructions uses the recommended [`git workflow`](https://kyma-project.io/community/contributing/#git-workflow-git-workflow) and the general [contribution flow](https://kyma-project.io/community/contributing/#contributing-rules-contributing-rules-contribute-code-or-content). Read also the [`CONTRIBUTING.md`](CONTRIBUTING.md) document that includes the contributing rules specific for this repository.

Follow these steps:

> **NOTE:** The example assumes you have the `$GOPATH` already set.

1. Fork the repository in GitHub.

2. Clone the fork to your `$GOPATH` workspace. Use this command to create the folder structure and clone the repository under the correct location:

    ```bash
    git clone git@github.com:{GitHubUsername}/kyma.git $GOPATH/src/github.com/kyma-project/kyma
    ```

    Follow the steps described in the [`git-workflow.md`](https://kyma-project.io/community/contributing#git-workflow-git-workflow-steps) document to configure your fork.

3. Install dependencies.

    Go to the main directory of the project in your workspace location and install the required dependencies:

    ```bash
    cd components/console-backend-service
    dep ensure -vendor-only
    ```

4. Build the project.

    Every project runs differently. Follow instructions in the main `README.md` document of the given project to build it.

5. Create a branch and start to develop.

    Do not forget about creating unit and acceptance tests if needed. For the unit tests, follow the instructions specified in the main `README.md` document of the given project. For the details concerning the acceptance tests, go to the corresponding directory inside the `tests` directory.
    Find the information on how to run changes on the cluster without a Docker image in the [Develop a service locally without using Docker](https://kyma-project.io/docs/root/kyma#tutorials-develop-a-service-locally-without-using-docker) document.

6. Test your changes.

    >**NOTE:** For more details about testing, go to the [Testing Kyma](https://kyma-project.io/docs/root/kyma#details-testing-kyma) document.

## Kyma users

Kyma is used by these companies:

<p align="center">
  <img src="https://github.com/kyma-project/website/blob/master/content/adopters/logos/sap.svg" alt="SAP" width="120" height="70" />
  <img src="https://github.com/kyma-project/website/blob/master/content/adopters/logos/accenture.svg" alt="Accenture" width="300" height="70" />
  <img src="https://github.com/kyma-project/website/blob/master/content/adopters/logos/netconomy.svg" alt="NETCONOMY" width="300" height="70" />
  <img src="https://github.com/kyma-project/website/blob/master/content/adopters/logos/digital_lights.svg" alt="Digital Lights" width="80" />
  <img src="https://github.com/kyma-project/website/blob/master/content/adopters/logos/FAIR_LOGO_HEADER.svg" alt="FAIR" width="80" />
  <img src="https://github.com/kyma-project/website/blob/master/content/adopters/logos/arithnea.svg" alt="ARITHNEA" width="80" />
</p>

Read about their [use cases](https://kyma-project.io/#used-by).
