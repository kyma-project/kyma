---
title: Use Compass
type: Tutorial
---

This tutorial present end-to-end use case scenario on how to enable and use Compass in Kyma.

## Prerequisites

?

## Steps

1. Install Kyma and enable Compass modules. Read [this](#installation-installation) document to learn how.   
2. Open Kyma Console and navigate to the Compass UI.
3. Compass UI

* Create a new Application. Your Application is by default added to the default scenario.
* Add API specification.
* Add Event specification.
* Assign default scenario to the existing Runtime

4. Runtime UI Applications view
* See that the Application is registered, two services are available
* Add a mapping to a given Namespace
* Go to the Service Catalog and 2 ServiceClasses should be available

5. Cleanup:
* Remove the mapping to the Namespace from the Runtime UI
* Compass UI, remove DEFAULT scenario from the Runtime
* The Application is removed from the Runtime UI.
