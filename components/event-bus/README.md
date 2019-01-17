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

## Publish controller and dispatcher Docker images of Nats Streaming provisioner on Knative using `ko`

[`ko`](https://github.com/google/go-containerregistry/tree/master/cmd/ko) makes it possible to publish an image for a Golang application without a Dockerfile. `ko` builds the application and creates a Docker image out of it. Follow these steps to publish Docker images for `controller` and `dispatcher` of Nats Streaming provisioner on Knative:

1. Install `ko`:

```
go get github.com/google/go-containerregistry/cmd/ko
```

2. Authenticate to Google Cloud and set project to `kyma-project`.

```
gcloud auth login
gcloud config set project kyma-project
```

3. Clone [knative/eventing](https://github.com/knative/eventing) and checkout the release branch you want to produce the images from.

4. Go to `natss` folder in knative/eventing:

```
cd eventing/pkg/provisioners/natss
```

5. Set `KO_DOCKER_REPO` to the repository and image path to be used.

```
export KO_DOCKER_REPO=eu.gcr.io/kyma-project/event-bus/knative/natss
```

6. Use `ko publish` to publish the Docker image for `controller`. `-t` is used to specify the tag, which should be the used Knative release branch. Setting `-B` disables the default behovior of adding MD5 hash after `KO_DOCKER_REPO`.

```
ko publish -B -t release-0.3 ./controller
```

7. Use `ko publish` to publish the Docker image for `dispatcher`.

```
ko publish -B -t release-0.3 ./dispatcher
```