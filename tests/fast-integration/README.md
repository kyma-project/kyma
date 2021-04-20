# Fast integration tests

## Overview

This project provides fast integration tests for Kyma. The goal is to decrease the minimal turnaround time ten times, from the current 90 minutes to less than 10 minutes. Fast integration tests will partially solve the problem of long-running tests. Other initiatives that are executed in parallel are equally important: switching to k3s, reducing Kubernetes provisioning time, and implementing the parallel installation of Kyma components.

The project also contains the `kyma-js` tool which you can use in the development process. `kyma-js` is a temporary solution that implements the subset of Kyma CLI commands, focusing on local development and minimal turnaround time (parallel installation). It provides some additional features that are useful for development, such as upgrading the selected components, skipping some components from the installation queue, testing the upgrade to the new eventing, and provisioning or deprovisioning a k3d cluster.

## Prerequisites

- [Node.js](https://nodejs.org) installation
- KUBECONFIG pointing to the Kubernetes cluster which has Kyma installed. If you don't have Kyma yet, you can quickly run it locally using [this project](https://github.com/kyma-incubator/local-kyma).
- Docker configured with 4GB RAM
- [k3d](https://github.com/rancher/k3d) - you can install it with these commands:
    ```bash
    brew install k3d
    ```
    or
    ```bash
    curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
    ```

- [crane](https://github.com/google/go-containerregistry/tree/master/cmd/crane) (optional) - a tool to copy Docker images. You can install it with this command:
    ```bash
    GO111MODULE=on go get -u github.com/google/go-containerregistry/cmd/crane
    ```


## Usage

To run tests locally, follow these steps:

1. Checkout the Kyma project:
```bash
git clone git@github.com:kyma-project/kyma.git
```

2. Install dependencies:
```bash
cd kyma/tests/fast-integration
npm install
```

3. Execute the tests:
```bash
npm test
```

## Local development

Here you have sample development tasks you can execute on your local machine working with the Kyma source code.

1. Install `kyma-js` as a global package:
    ```
    npm install -g kyma-js
    ```
    Another option is to navigate to the `fast-integration` folder, where the `kyma.js` file is located, and symlink the package folder:
    ```
    npm link
    ```
    The second option allows you to change the code of the installer and use it without building, publishing, and updating.

2. Create the local cluster:
    ```
    kyma-js provision k3d
    ```

3. Install Kyma without some modules and with the new eventing instead of Knative:
    ```
    kyma-js install -v --skip-components=monitoring,tracing,logging,kiali --new-eventing
    ```

4. Execute the Commerce Mock test with `DEBUG` enabled:
    ```
    DEBUG=true mocha test/2-commerce-mock.js
    ```

5. Upgrade some components:
    ```
    kyma-js install -v --component=application-connector --new-eventing
    ```

6. Delete the cluster and start from scratch:
    ```
    kyma-js deprovision k3d
    ```

To learn more about `kyma-js` possibilities, run:
```
kyma-js <command> --help
```

## FAQ

### Why don't you use Octopus?

[Octopus](https://github.com/kyma-incubator/octopus/blob/master/README.md) is a great tool for running tests inside the Kubernetes cluster in a declarative way. However, it is not the right tool for fast integration testing. The goal is to execute the tests in 4 minutes. With Octopus, you need 4 minutes or more before the tests even start (2 minutes to build the test image and push it to the Docker registry, 1 minute to deploy Octopus, and 1 minute to deploy the test Pod).

Octopus testing flow looks as follows:
1. Build a test image (or images) and push it (~ 2min/image).
2. Deploy Octopus (~1 min).
3. Deploy a test Pod (test image), (~ 1min/image).
4. In many tests, sleep 20 seconds to wait for a sidecar.
5. Deploy the "test scene" (~1 min/image).
6. Execute the test (5 sec/test).
7. Wait for the test completion and collect results (~1 min).

The fast-integration tests contain just two steps:
1. Deploy the "test scene" (~1-2 minutes, one scene for all the tests).
2. Execute the test (5 sec/test).

In this way, we can reduce testing phase from about 40 minutes to about 4 minutes.

### Why are tests written in Node.js and not in Go?

Tests are written in Node.js for several reasons:
- No compilation time
- Concise syntax (handling JSON responses from api-server or our test fixtures)
- Lighter dependencies (@kubernetes/client-node)
- Educational value for our customers who can read tests to learn how to use Kyma features (none of our customers write code in Go, they use JavaScript, Java, or Python)

### Which pipelines use fast-integration tests?

We have several pipelines that use fast-integration tests. See the list of pipelines with links to the Prow status page:

Pipeline | Description | Infrastructure
--|--|--|
[pre-main-kyma-integration-k3s](https://status.build.kyma-project.io/?job=pre-main-kyma-integration-k3s) | Job that runs on every PR before the merge to the `main` branch. | k3s
[post-main-kyma-integration-k3s](https://status.build.kyma-project.io/?job=post-main-kyma-integration-k3s) | Job that runs on every PR after it is merged to the `main` branch. | k3s
[kyma-integration-k3s](https://status.build.kyma-project.io/?job=kyma-integration-k3s) | Job that periodicially runs the fast-integration tests. | k3s
[kyma-integration-production-gardener-azure](https://status.build.kyma-project.io/?job=kyma-integration-production-gardener-azure) | Periodic job that tests the production profile in Kyma. | Gardener, Azure
[kyma-integration-evaluation-gardener-azure](https://status.build.kyma-project.io/?job=kyma-integration-evaluation-gardener-azure) | Periodic job that tests the evaluation profile in Kyma. | Gardener, Azure
