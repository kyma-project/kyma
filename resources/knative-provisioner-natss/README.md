# Knative Eventing NATS Streaming Provisioner

## Overview

This chart includes [knative nats streaming provisioner](https://github.com/knative/eventing/tree/master/contrib/natss/config) release files.

Included releases:
 * https://github.com/knative/eventing/blob/v0.4.1/contrib/natss/config/provisioner.yaml

Kyma-specific changes:

* The images changed to custom ones created from `kyma-incubator/eventing`.
* New environment variables: **{EB_USER}** and **{EB_PASS}** added for authentication.
* Environment variables **{DEFAULT_CLUSTER_ID}** and **{DEFAULT_NATSS_URL}** set for Kyma-specific needs.
* A new label `rand` added to Deployments to force Pod restart during the upgrade.
