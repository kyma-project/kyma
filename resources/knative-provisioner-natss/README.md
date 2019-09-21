# Knative Eventing NATS Streaming Provisioner

## Overview

This chart includes Knative NATS streaming provisioner release files.

Included releases:

* https://github.com/knative/eventing/tree/release-0.8/contrib/natss/config

Kyma-specific changes:

* New environment variables: **{EB_USER}** and **{EB_PASS}** added for authentication.
* Environment variables **{DEFAULT_CLUSTER_ID}** and **{DEFAULT_NATSS_URL}** set for Kyma-specific needs.
* A new label `rand` added to Deployments to force Pod restart during the upgrade.
