# Kubeless Test Client

## Overview

The `kubeless` folder contains tests for Kubeless functions. It contains the following:
* A [Dockerfile](Dockerfile) for the image used in Kyma Kubeless tests
* A [Go program](kubeless-tests.go), which executes the tests for the Kubeless chart
* The [ns.yaml](ns.yaml) file, which specifies the `kubeless-test` namespace
