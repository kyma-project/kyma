# Acceptance Tests

## Overview

This project contains the acceptance tests that you can run as part of Kyma testing process.
The tests are written in Go. Each component or group of scenarios has a separate folder with tests.

## Usage

To test your changes and build an image, use the `make build-image` command with **DOCKER_PUSH_REPOSITORY** and **DOCKER_PUSH_DIRECTORY** variables, for example:
```
DOCKER_PUSH_REPOSITORY=eu.gcr.io DOCKER_PUSH_DIRECTORY=/kyma-project/develop make build-image
```

### Add a new test

To add a new test:

1. Add a new package.
2. Modify the Dockerfile and the `build.sh` script to compile the test binary to `pkg.test`.
3. Add execution of the test to the `entrypoint.sh` script.
4. Add deletion of the binary to Makefile's cleanup step.

### Configure Kyma

After you build and push the Docker image, set the proper tag in the `resources/core/values.yaml` file, in the **acceptanceTest.imageTag** property.

### Run tests on Kyma

Follow these steps to run the acceptance tests on Kyma:

1. Update the test definition with a proper Docker image:
```bash
kubectl edit testdefinition -n kyma-system core
```

2. Create ClusterTestSuite to run the test:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: acceptance-test
spec:
  maxRetries: 1
  concurrency: 1
  selectors:
      matchNames:
      - name: core
        namespace: kyma-system
EOF
```
