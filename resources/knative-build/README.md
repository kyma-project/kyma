# Knative Build

## Overview

This module is still experimental.

Build is custom resource provided by knative. This can be used for building container images from the source code and push to docker registry of preference. This pushed docker image can be later pulled by the kantive-serving to start lambdas. For more details please refer to [knative build](https://github.com/knative/docs/tree/master/docs/build)

For leveraging full potential for knative builds refer the [samples](https://github.com/knative/build/tree/master/test)