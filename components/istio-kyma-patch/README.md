# Istio Kyma patch

## Overview

To run Kyma, the default Istio installation needs some changes. This application patches the already existing Istio 
installation so that Kyma can use Istio.

## Prerequisites

To run this application, install Istio in the` istio-system` Namespace.

## Usage

This section describes how to use the application.

### Configuration

The patch accepts the **CONFIG_DIR** environment variable which indicates a directory where patches are placed. If this 
variable is not set, the script uses the directory where it is placed as default.

### Details

The application performs several steps:
1. Check if the required CRDs are already deployed. The script reads the `required-crds` file which should contain a 
list of Istio's CRDs that Kyma requires. If CRDs are not deployed, the patch fails.

2. Configure the sidecar injector. The following changes are introduced to the `istio-sidecar-injector` ConfigMap:
    * The **policy** parameter is set to `disabled`.
    * The **zipkinAddress** points to Zipkin deployed in the `kyma-system` Namespace.
    * All containers have the default **resources.limits.memory** set to `128Mi` and **resources.limits.cpu** to `100m`.

3. Patch Istio components. The script looks for files in the `{resource-name}.{kind}.patch.json` format which contain a 
`JsonPatch`. The components are applied using the `kubectl patch` command. The modified resource may not exist, in which 
case the patch is skipped. See the [job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) to learn 
which patches are applied by default. On any other failure, such as wrong patch format or network error, the application 
fails.

4. Remove the unnecessary Istio components. The patch looks for the file named `delete` which should contain lines in 
the `{kind} {resource-name}` format. Every line describes an Istio resource which should be deleted from the 
`istio-system` Namespace. It is not an error if the resource to delete is missing. See the 
[job ConfigMap](../../resources/istio-kyma-patch/templates/configmap.yaml) to see which patches are applied by default.
