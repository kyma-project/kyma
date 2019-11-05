# Function Controller end-to-end test suite

Uses the [Kubernetes E2E framework](https://godoc.org/k8s.io/kubernetes/test/e2e/framework)

## Usage

`make test` runs all tests sequentially and assumes some in-cluster configuration is available to initialize Kubernetes
clients.

### Examples

Display all available test flags and exit.

```console
make test TEST_ARGS='-h'
```

Run locally, against the default cluster (context) configured in the referenced kubeconfig file. Equivalent to exporting
the **KUBECONFIG** environment variable.

```console
make test TEST_ARGS='-kubeconfig /home/me/.kube/config'
```

Enable Ginkgo's verbose mode, which prints log outputs immediately regardless of the outcome of tests.

```console
make test TEST_ARGS='-ginkgo.v'
```
