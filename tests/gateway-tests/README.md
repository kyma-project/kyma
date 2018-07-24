# Gateway acceptance tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Gateway testing process.
The tests are written in Go. Run them as standard Go tests.
Each component or group of scenarios has a separate folder, like `gateway`.

## Usage

This section provides information on building and versioning of the Docker image, as well as configuring the Kyma.


### Configuring the Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/values.yaml` file, in the`acceptanceTest.imageTag` property.

### Running locally

1. `helm ls` and pick a release you want to test.
```sh
$ helm ls
NAME               	REVISION	UPDATED                 	STATUS  	CHART                        	NAMESPACE
cluster-essentials 	1       	Tue May 22 10:47:14 2018	DEPLOYED	kyma-cluster-essentials-0.0.1	kyma-system
core               	2       	Tue May 22 11:00:03 2018	DEPLOYED	core-0.0.1                   	kyma-system
ec-default         	1       	Tue May 22 11:00:24 2018	DEPLOYED	gateway-0.0.1                	kyma-integration
hmc-default        	1       	Tue May 22 11:00:07 2018	DEPLOYED	gateway-0.0.1                	kyma-integration
istio              	1       	Tue May 22 10:47:21 2018	DEPLOYED	istio-0.5.1                  	istio-system
prometheus-operator	1       	Tue May 22 10:48:10 2018	DEPLOYED	prometheus-operator-0.18.1   	kyma-system
```

2. Let's say we pick `ec-default` release. In order to test it, run `helm test ec-default`.
```sh
$ helm test ec-default
RUNNING: test-ec-default-acceptance
PASSED: test-ec-default-acceptance
```

3. To see the logs, find the pod, that ran tests using `kubectl get po --show-all`. It'll be in `Completed` state.
```sh
$ kubectl get po --show-all
NAME                                  READY     STATUS      RESTARTS   AGE
ec-default-gateway-5dffff49f-drr7n    2/2       Running     0          1h
echo-service-745c674944-r67zj         1/1       Running     0          1h
hmc-default-gateway-669965894-5wwv8   2/2       Running     0          3h
test-ec-default-acceptance            0/1       Completed   0          7m # <<<<<< this one
```

and then just

```sh
kci logs test-ec-default-acceptance
=== RUN   TestGatewayHealth
=== RUN   TestGatewayHealth/SF_Gateway
--- PASS: TestGatewayHealth (0.00s)
    --- PASS: TestGatewayHealth/SF_Gateway (0.00s)
=== RUN   TestApiMetadata
--- PASS: TestApiMetadata (0.00s)
PASS
ok  	github.com/kyma-project/gateway-tests/test	0.003s
```
