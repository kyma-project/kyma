# Kiali

[Kiali](http://kiali.io) is a spyglass for the Istio Service Mesh. It provides the information on which microservices are a part of the specific Istio Service Mesh, how they are connected, and what is their performance status.

## Overview

This chart installs Kiali on a Kyma cluster.

## Prerequisites

- Install the Monitoring chart to make Kiali work.

## Details

### Installation

To install the chart with the release name `kiali` in the `kyma-system` Namespace, run:

    ```
    $ helm install kiali --name kiali --namespace kyma-system
    ```

### Unistallation

To uninstall/delete the `kiali` release but continue tracking the release, run:

    ```bash
    $ helm delete kiali
    ```

To uninstall/delete the `kiali` release completely and make its name free for further use, run:

    ```bash
    $ helm delete kiali --purge
    ```
