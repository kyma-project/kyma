<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-project/kyma/main/logo.png" width="235">
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/kyma)](https://goreportcard.com/report/github.com/kyma-project/kyma)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2168/badge)](https://bestpractices.coreinfrastructure.org/projects/2168)
[![Slack](https://img.shields.io/badge/slack-@kyma--community-yellow.svg)](http://slack.kyma-project.io)
[![Twitter](https://img.shields.io/badge/twitter-@kymaproject-blue.svg)](https://twitter.com/kymaproject)

## Overview

**Kyma** `/kee-ma/` is an application runtime that provides you a flexible and easy way to connect, extend, and customize your applications in the cloud-native world of Kubernetes. 

Out of the box, Kyma offers various functionalities, such as:  

- [Serverless](https://kyma-project.io/docs/kyma/latest/01-overview/serverless/) development platform to run lightweight Functions in a cost-efficient and scalable way
- [System connectivity]() that provides endpoint to securely register Events and APIs of external applications
- [Eventing](https://kyma-project.io/docs/kyma/latest/01-overview/eventing/) that provides messaging channel to receive events, enrich them, and trigger business flows using Functions or services
- [Service Mesh](https://kyma-project.io/docs/kyma/latest/01-overview/service-mesh/) for service-to-service communication and proxying
- [Service Management](https://kyma-project.io/docs/kyma/latest/01-overview/service-management/smgt-01-overview/) to use the built-in cloud services from such cloud providers as GCP, Azure, and AWS
- Secure API exposure
- In-cluster observability
- CLI supported by the intuitive UI through which you can connect your application to a Kubernetes cluster

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

## Contributing

Read the [`CONTRIBUTING.md`](CONTRIBUTING.md) document that includes the contributing rules and development steps specific for this repository.

## Kyma users

The following companies use Kyma:

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

## Join us

If you would like to join us and work together on the Kyma project, there are some prerequisite skills you should acquire beforehand. Git basic skills are the most important for a quick start with the code. Mastering Kubernetes skills is very important for your future work.

### Required programming skills

- Git basic skillset:
  - Forking a project from the `main` repository to your own repository
  - Checking out code from a public and private repository
  - Managing and fetching remote repositories
  - Creating a custom branch, adding and pushing commits to a remote branch of a forked project
  - Rebasing and merging a local branch with changes to the `main` branch
  - Creating and merging pull requests to the `main` branch
  - Interpreting automatic test results, rerunning a test suite
  - Resolving conflicts with the `main` branch

- Go basic skillset:
  - Installing and upgrading Go compiler
  - Setting up your IDE
  - Building a GoLang project
  - Running tests on a Golang project
  - Running code in the debug mode in your IDE
  - Understanding Makefiles and Dockerfiles
  - Downloading dependencies for the project
  - Understanding dependency tools such as `go mod` and `dep`
  - Downloading additional tools modules with the `go get` command

- Kubernetes basic skillset:
  - Understanding basic [Kubernetes architecture](https://shipit.dev/posts/kubernetes-overview-diagrams.html) and basic concepts such as: Namespace, Pod, Deployment, Secret, ConfigMap, ReplicaSet, Service, CustomResourceDefinition, Kubernetes Control Loop; understanding Kubernetes Design Patterns such as sidecars and init containers
  - Using a kubeconfig file to connect to a cluster
  - Browsing cluster resources using `kubectl` commands and editing Kubernetes resources using Terminal
  - Applying YAML files to a cluster with Kubernetes resources
  - Port forwarding from a running Pod to a local machine
  - Installing and using Minikube
  - Displaying logs from a container
  - Exporting Kubernetes objects to YAML files
  - Understanding Helm package manager
  - Certified Kubernetes Application Developer (CKAD) level preferred

- Docker basic skillset:
  - Listing all running Docker containers
  - Starting, stopping, deleting Docker containers
  - Exposing ports from running containers
  - Managing local image repositories
  - Pulling images from a remote repository and running them
  - Building images and tagging them
  - Pushing and managing images in your Docker Hub account
  - Executing `bash` commands inside containers

> **TIP:** Complete the [Docker and Kubernetes fundamentals](https://github.com/K8sAcademy/Fundamentals-HandsOn) training to get the basic Docker and Kubernetes knowledge.

- Cloud services skillset:
  - Logging in to Google Cloud Platform (GCP)
  - Understanding GCP basics concepts
  - Creating and deleting Kubernetes clusters in team projects on GCP
  - Creating Kubernetes shoot clusters on GCP and Azure

- Linux/Terminal basic skill set
  - Understanding basic `bash` scripting
  - Understanding the basics of the Unix filesystem
  - Performing basic operations on files (list, create, copy, delete, move, execute)
  - Sending REST queries with curl or HTTPie
  - CLI/Terminal confident use

- Fluency with command-line JSON and YAML processors, such as jq, yq, grep
- CI/CD experience (ideally Prow)

- Other skills
  - Understanding the Architecture Base Pattern
  - Understanding the Service Mesh concept
  - Basic Markdown editing


### Basic Kyma knowledge

These are the sources you can get the basic Kyma knowledge from:

- [Official Kyma documentation](https://kyma-project.io/)
- [Getting Started guide](https://kyma-project.io/docs/kyma/latest/02-get-started/)
- Kyma project [Youtube channel](https://www.youtube.com/watch?v=wqQflgmyboY&list=PLmZLSvJAm8FabPF4hLjScx-dDl84NK3l5)

### Open job positions

Kyma team is located mostly in Poland and Germany. See the open job positions for both locations:
- [Gliwice, Poland](https://jobs.sap.com/search/?createNewAlert=false&q=%23kymaopensource&optionsFacetsDD_department=&optionsFacetsDD_customfield3=&optionsFacetsDD_country=&locationsearch=)
- [Munich, Germany](https://jobs.sap.com/search/?createNewAlert=false&q=%23kyma&optionsFacetsDD_department=&optionsFacetsDD_customfield3=&optionsFacetsDD_country=&locationsearch=munich)

### FAQ

- **What is your IDE?**

  Nothing is enforced. People often use GoLand, Visual Studio Code, VIM.

- **How do you approach testing in Go? Do you use any frameworks?**

  We use tools such as classical Go runner, Gomega, Testify.

- **How to learn Go?**

  Here are some useful sources to learn Go:
  - [Official Go learning tutorials](https://go.dev/learn/)
  - [50 Shades of Go: Traps, Gotchas, and Common Mistakes for New Golang Devs](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/)
  - [Language converter](https://ide.onelang.io/?input=HelloWorldRaw) - this tool helps you to convert code from one language to any other one
