# Fast integration tests

## Overview

This project provides fast integration tests for Kyma. The goal is to decrease the minimal turnaround time from current 90 minutes to less than 10 minutes (ten times). Fast integration tests will solve the problem partially. Other initiatives that are executed in parallel are equally important: switching to k3s, starting to reduce Kubernetes provisioning time, and implementing parallel installation of Kyma components.

The current integration testing flow looks like this:
- Build test image(s), push it, ~ 2min/image
- Deploy octopus, ~1 min
- Deploy test pod (test image), ~ 1min/image
- Sleep 20 seconds to wait for sidecar (in many tests)
- Deploy "test scene", ~1 min/image
- Execute the test, 5 sec/test
- Wait for test completion and collect results. ~1 min

The plan is to keep only 2 steps:
- Deploy "test scene", 1-2 minutes (one scene for all the tests)
- Execute the test, 5 sec/test

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
Octopus is a great tool for running tests inside Kubernetes cluster in a declarative way. But it is not the right tool for fast integration testing. The goal is to execute the tests in 4 minutes. With Octopus, you need 4 minutes or more before test even start (2 minutes to build test image and push it to the Docker registry, 1 minute to deploy Octopus, and 1 minute to deploy the test pod).

### Why are tests written in node.js and not in Go?

For several reasons:
- no compilation time
- concise syntax (handling JSON responses from api-server or our test fixtures)
- lighter dependencies (@kubernetes/client-node)
- educational value for our customers who can read tests to learn how to use Kyma features (none of our customers write code in Go, they use JavaScript, Java or Python)