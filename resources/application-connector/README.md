# Application Connector

## Overview

The Application Connector connects an external solution to Kyma.

## Details

The Application Connector Helm chart contains all the global components:
- Metadata service
- Connector service

## Installation

The Application Connector is a part of the Kyma core and it installs automatically.

## Test Application Connector components on a local deployment

When you develop the components of the Application Connector, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `components/{COMPONENT_NAME}/scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh`, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

Alternatively, you can run only the `run-local-tests.sh` for the given component to build the image of the component's acceptance tests to the Minikube registry, create a Pod with the tests, and fetch the logs from that Pod.

This method of testing is available for the [Metadata Service](https://github.com/kyma-project/kyma/tree/master/components/metadata-service), the [Connector Service](https://github.com/kyma-project/kyma/tree/master/components/connector-service), and the [Remote Environment Controller](https://github.com/kyma-project/kyma/tree/master/components/remote-environment-controller).
