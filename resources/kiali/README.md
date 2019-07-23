# Kiali

[Kiali](http://kiali.io) is a spyglass for the istio service mesh and provides answers to the questions:
What microservices are part of my Istio service mesh? How are they connected? How are they performing?

## Introduction

This chart installs Kiali on a Kyma cluster.

## Prerequisites

- The `monitoring` chart must be installed in order to have Kiali functional
- The `jaeger` chart should be installed to have the Jaeger integration of Kiali functional
- The `logging` chart should be installed to have the Grafana integration of Kiali functional


## Installing the Chart

To install the chart with the release name `kiali` in namespace kyma-system call:

    ```
    $ helm install kiali --name kiali --namespace kyma-system
    ```

## Uninstalling the Chart

To uninstall/delete the `kiali` release but continue to track the release:
    ```
    $ helm delete kiali
    ```

To uninstall/delete the `kiali` release completely and make its name free for later use:
    ```
    $ helm delete kiali --purge
    ```
