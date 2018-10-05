# Istio Kyma patch

## Overview

To run Kyma, the default Istio installation needs some changes. This application patches the already existing Istio 
installation so that Kyma can use Istio.

## Usage

The application performs several steps:
1. Check if the required CRDs are already deployed. The script reads the `required-crds` file which should contain a 
list of Istio's CRDs that Kyma requires. If CRDs are not deployed patch will fail.

2. Configure the sidecar injector. The following changes are introduced to the `istio-sidecar-injector` ConfigMap:
    * The **policy** parameter is set to `disabled`.
    * The **zipkinAddress** points to Zipkin deployed in the `kyma-system` Namespace.
    * All containers have the default **resources.limits.memory** set to `128Mi` and **resources.limits.cpu** to `100m`.

3. Patch Istio components. The script looks for files in the `{resource-name}.{kind}.patch.json` format which contain a 
`JsonPatch`. The components are applied using the `kubectl patch` command. The modified resource may not exist. Patch 
will be skipped for such resource. On any other failure (e.g. wrongly formatted patch, network error, etc.) application 
will fail. See the [job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) to learn which patches are 
applied by default.

4. Remove the unnecessary Istio components. The patch looks for the file named `delete` which should contain lines in 
the `{kind} {resource-name}` format. Every line describes an Istio resource which should be deleted from the 
`istio-system` Namespace. It is not an error if the resource to delete is missing. See the 
[job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) to see which patches are applied by default.

Patch accepts following environmental variables:
* `CONFIG_DIR` which indicates a directory where patches are placed. If not set script will use directory it is placed 
in as default.

## Prerequisites

Istio must be installed in istio-system namespace in order to run this app.
