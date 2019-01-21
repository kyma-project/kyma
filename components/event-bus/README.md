# Event Bus

## Overview

The Event Bus enables Kyma to integrate with various other external solutions. The integration uses the `publish-subscribe` messaging pattern that allows Kyma to receive business Events from different solutions, enrich the events, and trigger business flows using lambdas or services defined in Kyma. See the [Event Bus documentation](https://github.com/kyma-project/kyma/tree/master/docs/event-bus/docs).

## Docker Images

Currently, Event Bus makes the following three Docker images available to the `kyma core` Helm chart:

- event-bus-publish
- event-bus-push
- event-bus-sub-validator

There are also end-to-end test Docker images to use as `helm tests`. See [the tests in the `event-bus` directory](https://github.com/kyma-project/kyma/tree/master/tests/event-bus) for more details.

## Development

The three binaries of `Event Bus` reside under `cmd/event-bus-XXXX` "e.g. `cmd/event-bus-publish`". They each have a Makefile to build and test the component as well as to create and push a Docker image. The following table explains the various make targets.


|Command| Description|
|-----------|------------|
|`make`|This is the default target for building the Docker image. It tests, compiles, creates, and appropriately tags a Docker image.|
|`make build`|Runs all the tests and the linter. It compiles the binary in the `bin` directory.|
|`make push`|Pushes the Docker image to the registry specified in the `REGISTRY` variable of the Makefile.|
|`make docker-build`|Creates a Docker image.|
|`make test`|Run all the tests.|
|`make vet`|Runs `go vet` on all sources including `vendor` but excluding the `generated` directory.|
|`make compile`|Builds a binary without running any tests.|

## Knative NATS Streaming provisioner

To use [NATS Streaming based provisioner](https://github.com/knative/eventing/tree/master/pkg/provisioners/natss), the controller and dispatcher images have to be available in GCR.

[`ko`](https://github.com/google/go-containerregistry/tree/master/cmd/ko) allows you to publish an image for a Golang application without a Dockerfile by building the application and creating a Docker image out of it. Perform the following steps to publish Docker images for the controller and dispatcher of Nats Streaming provisioner on Knative:

1. Install `ko`:

```
go get github.com/google/go-containerregistry/cmd/ko
```

2. Authenticate to Google Cloud Platform, set the project to `kyma-project`, and configure Docker:

```
gcloud auth login
gcloud config set project kyma-project
gcloud auth configure-docker
```

3. Clone [knative/eventing](https://github.com/knative/eventing) and check out the release branch you want to create the images from.

4. Go to the [`natss`](https://github.com/knative/eventing/tree/master/pkg/provisioners/natss) folder:

```
cd eventing/pkg/provisioners/natss
```

5. Export the repository and image path as the **KO_DOCKER_REPO** environment variable.

```
export KO_DOCKER_REPO=eu.gcr.io/kyma-project/event-bus/knative/natss
```

6. Run `ko publish` to publish the Docker image for the controller. Use the `-t` parameter to specify the Knative release branch tag. Add `-B` to disable the default adding of the MD5 hash after the **KO_DOCKER_REPO** variable.

```
ko publish -B -t release-0.3 ./controller
```

7. Run `ko publish` to publish the Docker image for the dispatcher.

```
ko publish -B -t release-0.3 ./dispatcher
```
