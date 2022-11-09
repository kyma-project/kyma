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

### Prepare a secret with cloud credentials to manage DNS.

Create the secret in the default namespace:

```
kubectl create secret google-credentials -n default --from-file=serviceaccount.json=serviceaccount.json
```

### Set the environment variables with custom domain

- `TEST_CUSTOM_DOMAIN` - set this environment variable with your desired custom domain.
- `TEST_DOMAIN` - set this environment variable with your installed by default Kyma domain.

After exporting these domains, run `make setup-custom-domain` to finish the default test setup.


### Run the tests

To start the test suite, run:

```
make test-custom-domain
```
