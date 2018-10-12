---
title: Local installation
type: Details
---

This document explains process of local installation. For better understanding of complex installation process, this guide describes it step by step.

## Start local installation

To fire local installation run following command:
```
./installation/cmd/run.sh
```

This script sets up default parameters, starts minikube, builds kyma-installer, generates local config, creates installer custom resource and finally sets up installer. Each step is explained in details below.

### Installation run parameters

Script `installation/cmd/run.sh` can be run with following option:

- `--skip-minikube-start` - it will not execute the `installation/scripts/minikube.sh` script, for details see section "Start minikube" in this document
- `--cr` - path to installer custom resource file, when not provided script will generate default local custom resource
- `--vm-driver` -  either `virtualbox` or `hiperkit` depending on your operating system

## Start minikube

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you don't need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.

Purpose of `installation/scripts/minikube.sh` is to configure and start minikube. The script also checks if your development environment is cofigured to handle kyma installation. This includes checking minikube and kubectl versions.In case minikube is already initialized you will be prompted to agree to remove previous minikube cluster. Script exits if you don't want to restart your cluster.

>**NOTE:** Minikube is configured to disable default nginx ingress controller.

>**NOTE:** For the complete list of parameters passed to `minikube start` command, refer to the `installation/scripts/minikube.sh` script.

Once minikube is up and running, the script adds develop domains to /ets/hosts.

## Build kyma-installer

Kyma installs locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/).

`installer` is application based on a [Kubernetes operator](https://coreos.com/operators/). Its purpose is to install helm charts defined via installer custom resource. `kyma-installer` is `installer` with bundled kyma charts. 

`installation/scripts/build-kyma-installer.sh` script will extract `kyma-installer` image name from the `installer.yaml` deployment file and use it to build docker image inside minikube. This image will contain local kyma sources from `resources` folder. 

>**NOTE:** For `kyma-installer` docker image details refer to the `kyma-installer/kyma.Dockerfile` file.

## Generate local configuration

TBD

## Create installer custom resource

In local installation scenario default `installer` custom resource will be used. For more details refer to `installation/resources/installer-cr.yaml.tpl`.