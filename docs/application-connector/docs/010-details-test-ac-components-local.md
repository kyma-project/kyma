---
title: Test Application Connector components on a local deployment
type: Details
---

When you develop the components of the Application Connector, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `resources/{COMPONENT_NAME}/scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh`, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

This script is available for the [Metadata Service](https://github.com/kyma-project/kyma/tree/master/components/metadata-service), the [Connector Service](https://github.com/kyma-project/kyma/tree/master/components/connector-service), and the [Remote Environment Controller](https://github.com/kyma-project/kyma/tree/master/components/remote-environment-controller).
