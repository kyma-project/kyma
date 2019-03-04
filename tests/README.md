# Tests

## Overview

The `tests` folder contains the sources for all Kyma tests.
A Kyma test is Pod, container, or image referenced in a Kyma module or chart test section. It provides the module's test functionality. 
A Kyma test runs against a running Kyma cluster. It ensures the integrity and functional correctness of the cluster with all installed modules. 
Each subdirectory in the tests directory defines sources for one test suite, usually focusing on one component. The resulting docker images are then referenced by the related Kyma modules or charts.

## Details

Every Kyma test resides in a dedicated folder containing its sources and a `README.md ` file. This file provides instructions on how to build and develop the test suite.

The test name, which is also the folder name, reflects the tested component without any prefix or suffix. For example,  `monitoring`.

The docker image resulting from the sources of a test suite resides in the tests subfolder.
Example: The Event Bus component has its acceptance tests in the `tests/event-bus` folder and produces the `XX/tests/event-bus:0.5.1` docker image.

Bundle the real e2e scenarios, such as **kubeless-integration** into one end-to-end subfolder. This folder contains one test project which executes all end-to-end tests divided into different packages by scenarios. For details, see the [`README.md`](end-to-end/README.md) file.
