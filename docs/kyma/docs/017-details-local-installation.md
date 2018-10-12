---
title: Local installation
type: Details
---

This document assumes that reader checked-out kyma repository. All scripts are fired and discussed in context of root path. 

## Start local installation

To fire local installation run following command:
```
./installation/cmd/run.sh
```

This script will set up default parameters, start minikube, build kyma-installer, generate local config, create installer custom resource and finally set up installer.

`DOMAIN_NAME` is set to `kyma.local`. For mac users virtual machine driver used to start minikube will be hyperkit and in other cases it will be `virtualbox`.

Script can be fired with following parameters:

- `--skip-minikube-start` - it will not fire `installation/scripts/minikube.sh` script, which means it will assume that minikube is up and runnng 
- `--cr` - path to installer custom resource file, when not provided script will generate default local custom resource
- `--vm-driver` -  by default `virtualbox` or `hyperkit` for mac users, for more details run `minikube start --help`

>**NOTE:** Please be aware that local installation does not support minikube start / stop scenario (we do not guarantee that all will work)

## Start minikube

Purpose of `installation/scripts/minikube.sh` is to configure and start minikube. To ensure installation script will check if proper version is installed. In next step version of `kubectl` will be checked. In last check minikube state will be checked. In case minikube is running you will be prompted to agree to remove previous minikube cluster. Script will be stopped if you don't want to restart your cluster.

>**NOTE:** Minikube is configure to disable default nginx ingress controller.

For the list of exact parameters passed to `minikube start` command please refer to `installation/scripts/minikube.sh`.

After minikube is up and running the script will add develop domains to /ets/hosts.

## Build kyma-installer

`installer` is application based on kubernetes operator patern. Its purpose is to install helm charts defined via installer custom resource. `kyma-installer` is `installer` with bundled kyma charts. 

`installation/scripts/build-kyma-installer.sh` script will extract `kyma-installer` image name from `installer.yaml` deployment and use it to build docker image inside minikube. This image will contain local kyma sources from `resources` folder. 

>**NOTE:** For `kyma-installer` docker image details refer to `kyma-installer/kyma.Dockerfile`.

## Generate local configuration

TBD

## Create installer custom resource

In local installation scenario default `installer` custom resource will be used. For more details refer to `installation/resources/installer-cr.yaml.tpl`.