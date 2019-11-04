# Knative Build

## Overview

Build is a custom resource provided by Knative. Use it to build container images from the source code and push them to docker registry of preference. You can later use knative-serving to pull this docker image to start lambdas. For details, see [knative build](https://github.com/knative/build/blob/master/README-old.md).

For details on how to leverage the full potential of knative-builds, see [samples](https://github.com/knative/build/tree/master/test).

This module is installing knative build crds.
