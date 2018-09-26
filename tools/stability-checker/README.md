# Stability Checker

## Overview

The Stability Checker runs the `testing-kyma.sh` script in a loop, and reports the results on a Slack channel.

## Installation

You can install the Stability Checker on the Kyma cluster as a chart. Find the chart definition in the `deploy/chart` directory.

> **NOTE:** You must install the chart after running the core tests, to avoid running the same tests in parallel.

## Usage

Stability Checker does not contain testing scripts. The chart value `.Values.pathToTestingScript` defines which script the system runs.
Ensure you have the file available in a persistent volume defined as `.Values.storage.claimName`, which is mounted as a `data` directory in the Stability Checker Pod.

To simulate the process of providing scripts, see the `local/provision_volume.sh` script, which populates the volume with files from the `local/input` directory.

### Deliver the chart

Download the gzipped chart from:

`https://github.com/kyma-project/stability-checker/raw/{branchName}/deploy/chart/stability-checker-0.1.0.tgz`

As another option, you can run the following:

```helm install https://github.com/kyma-project/stability-checker/raw/{branchName}/deploy/chart/stability-checker-0.1.0.tgz```

## Development
Use the following helpers for the local development:
- `./local_minikube_build.sh` which builds the Stability Checker Docker image on a Minikube registry.
- `./local/provision_volume.sh` which provisions a PersistentVolumeClaim (PVC) with testing scripts.
- `./local/charts/dummy` chart which contains simple and fast tests. To install it, execute the following command:
```
    helm install ./dummy --name dummy --namespace=kyma-system
```
- `./local_helm_install.sh` which installs the Stability Checker Helm chart with predefined values. 
The testing script points to the `testing-dummy.sh` which is a simplified version of `testing-kyma.sh`. The`dummy` chart is used in the `testing-dummy.sh` to speed up testing.