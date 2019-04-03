# Knative Eventing NATS Streaming Provisioner

## Overview

This chart includes [knative nats streaming provisioner](https://github.com/knative/eventing/tree/master/contrib/natss/config) release files.

Included releases:
 * https://github.com/knative/eventing/blob/v0.4.1/contrib/natss/config/provisioner.yaml

Kyma-specific changes:

* The images are changed to use the custom one we created from kyma-incubator/eventing.
* New environment variables `EB_USER` and `EB_PASS` are added for authentication.
* A new label `rand` is added to the Deployments to force Pod restart during the upgrade.