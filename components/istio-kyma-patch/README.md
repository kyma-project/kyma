# Istio Kyma patch

## Overview

To run kyma some changes needs to be made in default Istio isntallation. This application patches already existing Istio 
installation of Istio so it can be used by kyma.

The patch application performs several steps:
1. Check if required CRDs are deployed on setup. The script reads file named `required-crds`. The file should contain list of
Istio's CRDs which are required by kyma. 

1. The configuration of sidecar injector. The following changes are introduced to the `istio-sidecar-injector` ConfigMap:
    * The **policy** parameter is set to `disabled`.
    * The **zipkinAddress** points to zipkin deployed in `kyma-system` Namespace.
    * All containers have set default **resources.limits.memory** to `128Mi` and **resources.limits.cpu** to `100m`.

1. Istio components are patched. The script is expecting files in format `<resource-name>.<kind>.patch.json` which contain
patch in JsonPatch format. Components are patched using `kubectl patch` command. Patched resource may not exist. 
If other failure occurs the script must fail. See [job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) 
to see which patches are applied by default.

1. Unnecessary Istio components are removed - patch looks for file named `delete` which should contain lines in 
following format:

    ```<kind> <resource-name>```
    
    Every line must describe an Istio resource to delete. Resources are deleted from `istio-system` namespace. 
    It is not an error if resource to delete is missing. See [job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) 
    to see which patches are applied by default.

## Prerequisites

Istio must be installed in istio-system namespace in order to run this app.

## Usage

Patch accepts following environmental variables:
* `CONFIG_DIR` which indicates a directory where patches are placed. If not set script will use directory it is placed 
in as default.
