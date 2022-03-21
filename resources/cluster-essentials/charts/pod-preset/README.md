# Pod Preset

## Overview

The Pod Preset component allows injecting information such as Secrets, volume mounts, and environment variables during Pods creation.
Using a Pod Preset, you do not have to provide all information for every Pod.

## Details

This Helm chart installs the admission controller extracted from the main Kubernetes repository. Find the source code [here](https://github.com/jpeeler/podpreset-crd).
This Helm chart is required because the managed cluster does not enable the core Kubernetes admission controller for Pod Preset since it is in the alpha state.

This chart also provides the controller-manager which is responsible for restarting the Deployments whenever the Pod Preset changes.
This functionality is not present in the main Kubernetes implementation and by default is disabled in this Helm chart. If you want to enable the controller-manager, set the **controller.enabled** parameter to `true` in the [values.yaml](./values.yaml) file.
