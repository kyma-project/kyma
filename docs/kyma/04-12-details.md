---
title: Details
type: Installation
---

This section describes in details some of the installation mechanisms.

### Retry policy

Kyma Operator has a retry mechanism which makes sure that the installation does not fail in case of temporary issues like unavailability of Tiller or a prolonged process of creating a resource path for CustomResources in Kubernetes API Server. 

If the problem occurs during the installation of a component, the installation of that particular component is retried based on a configurable intervals between consecutive attempts. The default configuration consists of 5 retries with the following time intervals between each: 10s, 20s, 40s, 60s, 80s. After that, the installation is stopped and can be restarted manually by setting `action: install` label on the Installation CR. 

> **NOTE:** The configuration of retries can be adjusted by setting `backoffIntervals` argument in Installer Deployment. The value is a comma-separated list of numbers that represent intervals between consecutive retries. It also defines the total number of retries. For example, the default one is: `10,20,40,60,80`.

In case of the problem occurring before the installation of components, the installation will be retried every 30 seconds until successful. 
