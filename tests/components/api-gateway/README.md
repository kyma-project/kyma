# Api-gateway component tests

Api-gateway component tests use the [cucumber/godog](https://github.com/cucumber/godog) library.

## Prerequisites

- Kubernetes installed and kubeconfig configured to point to this cluster
- Kyma installed


### Environment variables

These environment variables determine how the tests are run on both Prow and your local machine:

- `EXPORT_RESULT` - set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON, and HTML report. The default value is `false`.

## Usage for standard api-gateway test suite

To start the test suite, run:

```
make test
```

If you don't have a cluster, you can run the tests on your local machine. To do so, run:

```
make test-k3d
```

This command creates a k3d cluster on your local machine, installs Kyma on it, and runs the tests.

## Usage for custom-domain test suite

### Set the environment variables with custom domain

If you are using Gardener cluster make sure your k8s cluster have cert & dns extensions. See [here](https://github.com/kyma-project/control-plane/issues/875)
Obtain a service account access key with permissions to maintain custom domain DNS entries and export it as json. See [here](https://cloud.google.com/iam/docs/keys-create-delete).

- `TEST_DOMAIN` - set this environment variable with your installed by default Kyma domain.
- `TEST_CUSTOM_DOMAIN` - set this environment variable with your custom domain.
- `TEST_SA_ACCESS_KEY_PATH` - set this environment variable with path to a service account access key, exported as a json.

### Run the tests

To start the test suite, run:

```
make test-custom-domain
```
