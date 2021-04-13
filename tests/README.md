# Tests

## Overview

The `tests` directory contains the sources for all Kyma tests.
A Kyma test is Pod, container, or image referenced in a Kyma module or chart test section. It provides the module's test functionality.
A Kyma test runs against a running Kyma cluster. It ensures the integrity and functional correctness of the cluster with all installed modules.
Each subdirectory in the tests directory defines sources for one test suite, usually focusing on one component. The resulting Docker images are then referenced by the related Kyma modules or charts.

## Details

Every Kyma test resides in a dedicated folder which contains its sources and a `README.md ` file. This file provides instructions on how to build and develop the test suite.

The test name, which is also the folder name, is the component's name without any prefix or suffix, such as `apiserver-proxy`.

The Docker image created from the sources of a test suite resides in a component folder marked with a suffix indicating the testing nature, such as `-integration-tests`.
Example: The apiserver-proxy component has its integration tests in `tests/integration/apiserver-proxy` folder and produces the `XX/apiserver-proxy-integration-tests:0.5.1` Docker image.

Bundle the real e2e scenarios, such as **kubeless-integration** into one `end-to-end` subfolder. This folder contains one test project which executes all end-to-end tests divided into different packages by scenarios.

Bundle integration tests, such as **apiserver-proxy** into one `integration` subfolder.

Bundle performance tests into one `perf` subfolder.

## Development

Follow [this](https://github.com/kyma-project/kyma/blob/main/resources/README.md) development guide when you add a new test to the `kyma` repository.
