<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-project/kyma/main/logo.png" width="235">
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/kyma)](https://goreportcard.com/report/github.com/kyma-project/kyma)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2168/badge)](https://bestpractices.coreinfrastructure.org/projects/2168)
[![Slack](https://img.shields.io/badge/slack-@kyma--community-yellow.svg)](http://slack.kyma-project.io)
[![Twitter](https://img.shields.io/badge/twitter-@kymaproject-blue.svg)](https://twitter.com/kymaproject)

## Overview

**Kyma** `/kee-ma/` is a platform for extending applications with microservices and [serverless](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/serverless/) Functions. It provides CLI and UI through which you can connect your application to a Kubernetes cluster. You can also expose the application's API or events securely thanks to the built-in [Application Connector](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/application-connectivity/). You can then implement the business logic you require by creating microservices or serverless Functions. Trigger them to react to particular events or calls to your application's API.

To limit the time spent on coding, use the built-in cloud services from [Service Management](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/service-management/smgt-01-overview/) from such cloud providers as GCP, Azure, and AWS.

<p align="center">
<a href="https://youtu.be/kP7mSELIxXw">
<img src="./docs/kyma/assets/withoutprov4.gif" style="max-width:100%;">
</a>
</p>

Go to the [Kyma project website](https://kyma-project.io/) to learn more about our project, its features, and components.

## Installation

Install Kyma locally or on a cluster. See the [Installation guides](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/02-install-kyma/) for details.

> **NOTE:** Make sure to install the latest Kyma version and keep it up to date by [upgrading Kyma](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/05-upgrade-kyma/).

## Usage

Kyma comes with the ready-to-use code snippets that you can use to test the extensions and the core functionality. See the list of existing examples in the [`examples`](https://github.com/kyma-project/examples) repository.

## Development

Develop on your remote repository forked from the original repository in Go.
Read also the [`CONTRIBUTING.md`](CONTRIBUTING.md) document that includes the contributing rules specific for this repository.

Follow these steps:

> **NOTE:** The example assumes you have the `$GOPATH` already set.

1. Fork the repository in GitHub.

2. Clone the fork to your `$GOPATH` workspace. Use this command to create the folder structure and clone the repository under the correct location:

    ```bash
    git clone git@github.com:{GitHubUsername}/kyma.git $GOPATH/src/github.com/kyma-project/kyma
    ```

    Follow the steps described in the [`git-workflow.md`](https://kyma-project.io/community/contributing/03-git-workflow/) document to configure your fork.

3. Build the project.

    Every project runs differently. Follow instructions in the main `README.md` document of the given project to build it.

4. Create a branch and start to develop.

    Do not forget about creating unit and acceptance tests if needed. For the unit tests, follow the instructions specified in the main `README.md` document of the given project. For the details concerning the acceptance tests, go to the corresponding directory inside the `tests` directory.

5. Test your changes.

## Kyma users

[Read](https://kyma-project.io/#used-by) how these companies use Kyma:

<p align="center">
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/sap.svg" alt="SAP" width="120" height="70" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/accenture.svg" alt="Accenture" width="300" height="70" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/netconomy.svg" alt="NETCONOMY" width="300" height="70" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/neteleven.svg" alt="neteleven" width="250" height="60" />
</p>
<p align="center">
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/arithnea.svg" alt="ARITHNEA" width="300" height="130" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/digital_lights.svg" alt="Digital Lights" width="200" height="130" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/FAIR_LOGO_HEADER.svg" alt="FAIR" width="300" height="130" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/Sybit-Logo.svg" alt="FAIR" width="300" height="130" />
</p>
<p align="center">
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/arineo.svg" alt="Arineo" width="250" height="40" />
  <img src="https://github.com/kyma-project/website/blob/main/content/adopters/logos/dotsource.svg" alt="dotSource" width="250" height="40" />
</p>
