---
title: Error handling
type: Installation
---

Kyma Operator features a retry mechanism to handle temporary issues such as prolonged process of creating a resource path for custom resources in the Kubernetes API Server. 
If an error occurs while processing a component, Kyma Operator restores the initial state of that component and retries the step. Specific behavior of the controller depends on the nature of the operation that was interrupted by an error.

## Installation error

If an error occurs during component installation, Kyma Operator deletes the corresponding Helm release and retries the operation. If such a release does not exist, the deletion step is skipped.

## Upgrade error

If an error occurs during component upgrade, Kyma Operator rolls back the corresponding Helm release to the last deployed revision and retries the operation. If the release history does not include a deployed revision, the controller returns an error and stops the process. 

## Retry policy
 
The retry policy is based on the configuration that specifies the intervals between consecutive attempts. The default configuration consists of 5 retries with the following time intervals between each attempt: 10s, 20s, 40s, 60s, 80s. After that, the installation is stopped and can be restarted manually by setting the `action: install` label on the Installation CR. 

> **NOTE:** The configuration of retries can be adjusted by setting the `backoffIntervals` argument in the Installer Deployment. The value is a comma-separated list of numbers that represent the intervals between consecutive retries. It also defines the total number of retries. For example, the default one is: `10,20,40,60,80`.

If an error occurs before the installation of components, the installation is retried every 30 seconds until successful.
