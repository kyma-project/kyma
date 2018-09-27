# Istio Kyma patch

## Overview

Application performs patch of already existing istio installation so kyma can be run.

## Configuration

Patch accepts following environmental variables:
* `CONFIG_DIR` - directory where patches are placed. If not set script will look in directory it is placed in

## Process

Patch consists of several steps:
1. Configuration of sidecar injector - istio-sidecar-injector configmap has following changes:
    * `policy` is set to `disabled`
    * `zipkinAddress` is changed so it points to zipkin deployed in kyma-system namespace
    * all containers have set default `resources.limits.memory` to `128Mi` and `resources.limits.cpu` to `100m`

1. Check if required CRDs are deployed on setup - script is looking for file named `required-crds`. In each line of this 
file should contain one name of CRD which kyma requires to start.

1. Istio components are patched - patch is expecting files in format `<resource-name>.<kind>.patch.json` which contain
patch in JsonPatch format. Patch allows that resource to patch does not exist.
See [job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) to see which patches are applied by 
default.

1. Unnecessary istio components are removed - patch looks for file named `delete` which should contain lines in 
following format:

    ```<kind> <resource-name>```
    
    Every line must describe an istio resource to delete. Resources are deleted from `istio-system` namespace. 
    Patch allows that resource to delete does not exist. See [job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) 
    to see which patches are applied by default.