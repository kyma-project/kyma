# Event Bus

## Overview

The Event Bus enables Kyma to integrate with various other external solutions. The integration uses the `publish-subscribe` messaging pattern that allows Kyma to receive business Events from different solutions, enrich the events, and trigger business flows using lambdas or services defined in Kyma. See the [Event Bus documentation](https://kyma-project.io/docs/components/event-bus/).

## Docker Images

Currently, Event Bus makes the following three Docker images available to the `kyma core` Helm chart:

- event-bus-publish-knative
- event-bus-subscription-controller-knative

There are also end-to-end test Docker images to use as `helm tests`. See [the tests in the `event-bus` directory](https://github.com/kyma-project/kyma/tree/master/tests/event-bus) for more details.

## Development

The two binaries of `Event Bus` reside under `cmd/event-bus-XXXX` "e.g. `cmd/event-bus-publish-knative`". They each have a Makefile to build and test the component as well as to create and push a Docker image. The following table explains the various make targets.


|Command| Description|
|-----------|------------|
|`make`|This is the default target for building the Docker image. It tests, compiles, creates, and appropriately tags a Docker image.|
|`make build`|Runs all the tests and the linter. It compiles the binary in the `bin` directory.|
|`make push-image`|Pushes the Docker image to the registry specified in the `REGISTRY` variable of the Makefile.|
|`make build-image`|Creates a Docker image.|
|`make test`|Run all the tests.|
