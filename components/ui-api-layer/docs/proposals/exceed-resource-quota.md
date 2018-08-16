# Check when the ResourceQuota is exceeded

[ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/) provides constraints that limit aggregate resource consumption per Namespace.

The ResourceQuota values can be as follows:

```bash
Name:            kyma-default
Resource         Used       Hard
--------         ---        ---
limits.memory    7348440Ki  10Gi
requests.memory  3727576Ki  7Gi
```

The Console calls for the ResourceQuota status. The calls are triggered after: 
- switching the environment
- uploading a resource
- creating a lambda
- opening environment's details
 
The idea is to create an endpoint which returns the ResourceQuotaStatus.
The ResourceQuotaStatus contains the flag and the optional message. The value of this flag informs you if the ResourceQuota is exceeded.

## Multistage check

To check if you exceeded the ResourceQuota in the Environment, you must implement a few steps.

### Check the status of the ResourceQuotas

List ResourceQuotas in the given Namespace and loop through them. You exceeded the ResourceQuota if at least one of the `.status.used` values equals or is bigger than its equivalent from the `.spec.hard` value.
In this situation, return the ResourceQuotaStatus with a flag set to `true`.

You must specify which resources you want to check.
For Environments, check the following resources:
- memory requests and limits
- CPU requests and limits
- number of Pods

### Check the ReplicaSets

List ReplicaSets in the given Namespace and loop through them. You must check every ReplicaSet which does not reach the number of desired replicas.
First of all, check if the ReplicaSet has an **OwnerReference** to the Deployment. If yes, you must get the Deployment's `.spec.strategy.rollingUpdate.maxUnavailable` value.
Parse that value to **int** and use it to decrease the number of desired replicas in the `if` statement which checks if the number of desired replicas is reached.

>**NOTE:** The value of `.spec.strategy.rollingUpdate.maxUnavailable` can be expressed as a percentage or a number. Find more information [here](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#max-unavailable).

If the desired number of replicas is not reached, check if any ResourceQuota blocks the progress of the ReplicaSet.
To achieve that, implement logic which calculates if the ReplicaSet has enough memory to create the next replica.
To calculate how many resources the ReplicaSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`. It gives you the total amount of resources available to use in the given Namespace.

Loop through the ResourceQuota list and check if any ResourceQuota blocks the progress of the ReplicaSet by comparing the resources available in the Environment and the resources required to create the next replica.
If there is not enough resources to create the next replica and the number of desired replicas is not reached, return a ResourceQuotaStatus with a flag set to `true`.

>**NOTE:** The Kubeless functions work with this solution because they use the Deployment and the ReplicaSet underneath.

### Check the StatefulSets

List StatefulSets and loop through them.
For each StatefulSet that has a number of replicas lower than expected, implement the logic which checks if any ResourceQuota blocks the progress of the StatefulSet.
To achieve that, implement the logic which calculates if the StatefulSet has enough memory to create the next replica.
To calculate how many resources the StatefulSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`. It gives you the total number of resources available to use in the given Namespace.

Loop through the ResourceQuota list and check if any ResourceQuota blocks the progress of the ReplicaSet by comparing the resources available in the Environment and the resources required to create the next replica.
If there is not enough resources to create the next replica and the number of replicas is not reached, return a ResourceQuotaStatus with a flag set to `true`.

### End of checking

After all checks passed and the resources usage in the given Environment did not exceed the ResourceQuota, return the ResourceQuotaStatus with a flag set to `false`.

[Here](https://github.com/dtaylor113/origin-web-console/blob/master/app/scripts/services/quota.js#L86) you can find the example implementation of the checking logic in UI.


## Percentage usage of the ResourceQuota

>**NOTE:** The Kyma team rejected that solution for now but it may be implemented in the future.

The percentage thresholds are stored in the ConfigMap. Each resource type has its own threshold.
In the ConfigMap, there is also one default percentage threshold for all resource types which do not have a threshold set.

In the future, the user will be able to configure thresholds stored in the ConfigMap.
With a percentage approach, you can see the actual usage for every environment on the UI.

The UI must mention the user when he exceeded the percentage threshold for a given resource.

To avoid the situation when the user exceeded the ResourceQuota without realizing it, the UI should send a notification every time any ReplicaSet in the Namespace does not reach the desired number of Pods.
