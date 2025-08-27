# Docker Registry Module Configuration

## Overview

The Docker Registry module has its own operator (Docker Registry Operator). It watches the Docker Registry custom resource (CR) and reconfigures (reconciles) the Docker Registry workloads.

The DockerRegistry CR allows you to store images in five ways: filesystem, Azure, s3, GCP, and BTP Object Store, each requiring specific configurations. See [Reigstry Storage Configuration](00-30-storage-configuration.md) to learn more.

The Docker Registry CR becomes an API to configure the Docker Registry module. You can't configure anything right now, but you will be able to do so soon.

The default configuration of the Docker Registry module is the following:

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: DockerRegistry
   metadata:
     name: default
     namespace: kyma-system
   spec: {}

   ```
