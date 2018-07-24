# Pod State Watcher

## Overview
This tool inspects the state of all Pods.
If at least one Pod is not in the running state or restarts constantly, the program exits with the code other than `0`.

The program runs outside of the cluster so it requires the `kubeconfig` file to access the cluster.


## Prerequisites

The `build` script requires golint and Go in 1.9 version.

## Usage
This section describes how to build a project image. It also provides the parameters to start and configure a container from the previously built image.

### Build a project image

Execute the **build.sh** script which performs tests and static code analysis, and creates the `watch-pods` docker image.

### Configuration options
The following flags are accessible to start and configure the previously built binary file:
- **kubeconfig** (default: in cluster) - a path to the kubeconfig file
- **minWaitingPeriod** (default: 1 minute) - time needed for the cluster to stabilize, after which tests are performed
- **maxWaitingPeriod** (default: 3 minutes) - the maximum period of time in which test are performed
- **reqStabilityPeriod** (default: 1 minute) - the required stability period which informs for how long the container cannot be restarted to be considered as stable

### Run the image
```
docker run --rm --env "ARGS=-maxWaitingPeriod=10m -ignorePodsPattern='sf-core-azure-broker-docs-*'" watch-pods:latest
```
