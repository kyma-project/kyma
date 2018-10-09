# Pod Preset

## Overview

The Pod Preset component allows injecting information such as Secrets, volume mounts, and environment variables during Pods creation.
Using a PodPreset, you do not have to provide all information for every Pod.

## Details

This Helm chart installs the admission controller extracted from the main Kubernetes repository. Find the source code [here](https://github.com/jpeeler/podpreset-crd).
This Helm chart is required because the managed cluster does not enable the core Kubernetes admission controller for PodPreset since it is in the alpha state.

This chart also provides the controller-manager which is responsible for restarting the Deployments whenever the PodPreset changes.
This functionality is not present in the main Kubernetes implementation and by default is disabled in this Helm chart. If you want to enable the controller-manager, set the **controller.enabled** parameter to `true` in the [values.yaml](./values.yaml) file.

### Images
The Docker images are pushed to the [Docker Hub](https://hub.docker.com/) under the [mszostok](https://hub.docker.com/u/mszostok/) repository. This is a temporary solution until the official images are pushed from the [podpreset-crd](https://github.com/jpeeler/podpreset-crd) repository.

The `mszostok/service-catalog-podpreset-controller` Docker image was build by executing the `make docker-build` command from the [podpreset-crd](https://github.com/jpeeler/podpreset-crd) repository. The `0.0.2` version is build from [this](https://github.com/jpeeler/podpreset-crd/commit/4d6e1a45dd59ac149a171f43dc1499d3ee4901a8) commit.
The `mszostok/service-catalog-admission-webhook` Docker image was build by executing the `make docker-build-webhook` command from the [podpreset-crd](https://github.com/jpeeler/podpreset-crd) repository. The `0.0.2` version is build from [this](https://github.com/jpeeler/podpreset-crd/commit/4d6e1a45dd59ac149a171f43dc1499d3ee4901a8) commit.
