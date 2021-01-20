# Fast integration tests

## Overview

This project provides fast integration tests for Kyma. The goal is to decrease the minimal turnaround time from the current 90 minutes to less than 10 minutes (ten times). Fast integration tests will solve the problem partially. Other initiatives that are executed in parallel are equally important: switching to k3s, reducing Kubernetes provisioning time, and implementing the parallel installation of Kyma components.

The current integration testing flow looks like this:
1. Build a test image (or images) and push it (~ 2min/image).
2. Deploy [Octopus](https://github.com/kyma-incubator/octopus/blob/master/README.md) (~1 min).
3. Deploy a test Pod (test image), (~ 1min/image).
4. In many tests, sleep 20 seconds to wait for a sidecar.
5. Deploy the "test scene" (~1 min/image).
6. Execute the test (5 sec/test).
7. Wait for the test completion and collect results (~1 min).

The plan is to keep only 2 steps:
1. Deploy the "test scene" (~1-2 minutes, one scene for all the tests).
2. Execute the test (5 sec/test).

In this way, we can reduce testing phase from about 40 minutes to about 4 minutes.

## Prerequisites

- A [node.js](https://nodejs.org) installation.
- KUBECONFIG pointing to the Kubernetes cluster which has Kyma installed. If you don't have Kyma yet, you can quickly run it locally using [this project](https://github.com/kyma-incubator/local-kyma).


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

## FAQ

### Why don't you use Octopus?
Octopus is a great tool for running tests inside the Kubernetes cluster in a declarative way. But it is not the right tool for fast integration testing. The goal is to execute the tests in 4 minutes. With Octopus, you need 4 minutes or more before the tests even start (2 minutes to build the test image and push it to the Docker registry, 1 minute to deploy Octopus, and 1 minute to deploy the test Pod).

### Why are tests written in node.js and not in Go?

Tests are written in node.js for several reasons:
- No compilation time
- Concise syntax (handling JSON responses from api-server or our test fixtures)
- Lighter dependencies (@kubernetes/client-node)
- Educational value for our customers who can read tests to learn how to use Kyma features (none of our customers write code in Go, they use JavaScript, Java, or Python)

### Which pipelines use fast-integration tests?

We have several pipelines that use fast-integration tests. See the list of pipelines with links to the Prow status page:

Pipeline | Description | Infrastructure
--|--|--|
[pre-master-kyma-integration-k3s](https://status.build.kyma-project.io/?job=pre-master-kyma-integration-k3s) | Job that runs on every PR before the merge to the `master` branch. | k3s
[post-master-kyma-integration-k3s](https://status.build.kyma-project.io/?job=post-master-kyma-integration-k3s) | Job that runs on every PR after it is merged to the `master` branch. | k3s
[kyma-integration-k3s](https://status.build.kyma-project.io/?job=kyma-integration-k3s) | periodic job | k3s
[kyma-integration-production-gardener-azure](https://status.build.kyma-project.io/?job=kyma-integration-production-gardener-azure) | Periodic job that tests the production profile in Kyma. | Gardener, Azure
[kyma-integration-evaluation-gardener-azure](https://status.build.kyma-project.io/?job=kyma-integration-evaluation-gardener-azure) | Periodic job that tests the evaluation profile in Kyma. | Gardener, Azure
