# Istio component tests

Istio component tests use the [cucumber/godog](https://github.com/cucumber/godog) library.

## Prerequisites

- Kubernetes installed and kubeconfig configured to point to this cluster
- Kyma installed with at least Istio component
- Environment variables exported (the only required environment variable is `KYMA_PROFILE`)

### Environment variables

These environment variables determine how the tests are run on both Prow and your local machine:

- `KYMA_PROFILE` - set this environment variable accordingly to the Kyma profile installed on the Kubernetes cluster. The possible values are `evaluation` or `production`.
- `EXPORT_RESULT` - set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON, and HTML report. The default value is `false`.

## Usage

### Start the test suite

Having met all the requirements simply run:

```make test```

#### Prepare cluster on your local machine

We have also provided a Make target that will create k3d cluster, install Kyma on it and run the tests:

```make test-k3d```

The steps are as follows:

1. Provision a k3d cluster on your local machine:

```make provision-k3d```

2. Install Kyma with the Istio component on your cluster:

```make kyma-istio-deploy```

3. Run the tests:

```make test```