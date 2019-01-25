# Ark plugins

## Overview

Ark plugins provide an implementation of a few plugins required to properly restore the Kyma cluster and resources created inside it. They focus mainly on resources related to the Service Catalog, such as instances or bindings. Each plugin is defined in a separate file inside the [`internal/plugins`](internal/plugins) folder. The purpose of each plugin is defined in the file's comments.

The structure of folders and files is based on the [Ark plugin example repository](https://github.com/heptio/ark-plugin-example).

## Installation

Run a Docker image build from the [Dockerfile](Dockerfile) as an init container inside the deployment of the [Ark chart](../../resources/ark). During development, you can push a built image to your own image repository and run:

```bash
ark plugin add {yourRepo/imageName:tag}
```
