# Tests

## Overview

The tests folder contains the sources for all Kyma tests.
A Kyma test is Pod/container/image referenced in a Kyma module/chart test section to provide the module's test functionality. 
A Kyma test gets executed against a running Kyma cluster to assure integrity and functional correctness of the cluster with all modules installed. These are acceptance tests.
Each subfolder in the tests directory defines sources for one test suite, usually focusing on one component. The resulting docker images are then referenced by the related Kyma modules/charts.

## Details

Every Kyma test has a dedicated folder containing its sources and a README.md containing further instructions on how to build and develop the test suite.

The test name and with that the folder name should reflect the component under test. It should not have a prefix and no suffix, just the component name under test, for example _monitoring_.

The docker image resulting from the sources of a test suite should reside in the tests subfolder.
Example: The Event-Bus component has its acceptance tests in the tests/event-bus folder and produces the XX/tests/event-bus:0.5.1 docker image.

Bundle the real e2e scenarios (like kubeless-integration) into one end-to-end subfolder. Here we should have one test project which executes all end-to-end tests divided by scenarios into different packages, see the [README.md](end-to-end/README.md) for more details.