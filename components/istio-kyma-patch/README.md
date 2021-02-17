# Istio Kyma patch

## Overview

To run Kyma, the default Istio installation needs some changes. This application patches the already existing Istio 
installation so that Kyma can use Istio.

## Prerequisites

To run this application, install Istio in the `istio-system` Namespace.

## Usage

This section describes how to use the application.

### Configuration

The patch accepts the **CONFIG_DIR** environment variable which indicates a directory where patches are placed. If this 
variable is not set, the script uses the directory where it is placed as default.

### Details

The application performs several steps:
1. Check if the required CRDs are already deployed. The script reads the `required-crds` file which must contain a 
list of Istio CRDs that Kyma requires. If any of these CRDs is not deployed, the patch fails.

2. Configure the sidecar injector. The `istio-sidecar-injector` ConfigMap has these modifications:
    * The **policy** parameter is set to `disabled`.
    * The **zipkinAddress** points to Zipkin deployed in the `kyma-system` Namespace.
    * All containers have the default **resources.limits.memory** set to `128Mi` and **resources.limits.cpu** to `100m`.

3. Patch Istio components. The script looks for `{resource-name}.{kind}.patch.json` format files which contain a 
`JsonPatch`. The components are applied using the `kubectl patch` command. The modified resource may not exist, in which 
case the patch is skipped. In case of failure, such as wrong a patch format or a network error, the 
application fails.

4. Remove the unnecessary Istio components. The patch acts on a delete file which contains lines that follow the 
`{kind} {resource-name}` format. Every line points to an Istio resource which the patch removes from the
`istio-system` Namespace. The system doesn't return an error if a resource listed in the delete file is not present in 
the istio-system Namespace.

5. Enable default sidecar injection. By default, sidecar injection is enabled in the Namespaces labeled with `istio-injection: enabled`. The patch reverses this behavior: sidecar injection is enabled in all Namespaces, except those labeled with `istio-injection: disabled`.

6. Label the Namespaces that should not allow sidecar injection. The list of excluded Namespaces is declared under the **injection-in-namespaces** key in the `istio-kyma-patch-config` Configmap.
