# Ark plugins

## Overview

Ark plugins provide an implementation of a few plugins required to properly restore kyma cluster and created resources inside it. It's mainly focused on resources related to the service catalog (instances, bindings, etc). Each plugin is defined in separate file inside [internal/plugins](internal/plugins) folder. Purpose of each plugin is defined in the file's comments.

Structure of the folders/files is based on [ark plugin example repository](https://github.com/heptio/ark-plugin-example).

## Installation

The docker image build from [Dockerfile](Dockerfile) is run as an init container inside deployment of [ark chart](../../resources/ark). During development, you can push built image to your own image repository and run `ark plugin add [yourRepo/imageName:tag]`.