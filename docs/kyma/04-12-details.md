---
title: Details
type: Installation
---

This section describes in details some of the installation mechanisms.

### Retry policy

Kyma Operator has a retry mechanism which makes sure that the installation does not fail in the case of temporary issues, such as Tiller unavailability or a prolonged process of creating a resource path for custom resources in the Kubernetes API Server. 

If an error occurs during the installation of a component, the installation of that particular component is retried based on the provided configuration that specifies the intervals between consecutive attempts. The default configuration consists of 5 retries with the following time intervals between each attempt: 10s, 20s, 40s, 60s, 80s. After that, the installation is stopped and can be restarted manually by setting the `action: install` label on the Installation CR.

> **NOTE:** The configuration of retries can be adjusted by setting the `backoffIntervals` argument in the Installer Deployment. The value is a comma-separated list of numbers that represent the intervals between consecutive retries. It also defines the total number of retries. For example, the default one is: `10,20,40,60,80`.

In the case of an error occurring before the installation of components, the installation is retried every 30 seconds until successful. 
