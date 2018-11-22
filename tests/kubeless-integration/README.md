# Kubeless Test Client

## Overview

The `kubeless-integration` folder contains tests for Kubeless integration with other Kyma resources. It contains the following:
* A [Dockerfile](Dockerfile) for the image used in Kyma Kubeless tests
* A [Go program](test-kubeless.go), which executes the tests for the Kubeless chart
* The [ns.yaml](ns.yaml) file, which specifies the `kubeless-test` namespace
* The JavaScript files of `test-event` and `test-hello`
* The [svc-instance.yaml](svc-instance.yaml) file deploys the Redis service instance and service binding.
