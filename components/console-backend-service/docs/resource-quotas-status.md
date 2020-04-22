# ResourceQuotasStatus

[ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/) provides constraints that limit the aggregate resource consumption per Namespace.

The example ResourceQuota values look as follows:
```bash
Name:            kyma-default
Resource         Used       Hard
--------         ---        ---
limits.memory    7348440Ki  10Gi
requests.memory  3727576Ki  7Gi
```
When the `used` values of the ResourceQuota exceed the `hard` values, the resource creation in the Namespace is blocked.

The ResourceQuotasStatus contains the flag and the list of exceeded ResourceQuotas limits, together with a set of resources which exceed that limits. The value of the flag informs you if any ResourceQuota is exceeded in any possible way.

The ResourceQuotasStatus detects if any ReplicaSet or StatefulSet in your Namespace is blocked by the ResourceQuota and therefore cannot progress.
To check if any ReplicaSet or StatefulSet is blocked, calculate the required number of resources to create another replica. If there are not enough resources to create another replica, return the ResourceQuotaStatus with the flag set to `true`.

The Console calls for the ResourceQuotasStatus automatically. The calls are triggered after:
- switching the Namespace
- uploading a resource
- creating a function 
- opening Namespace's details

## Implementation

This section contains the steps of the ResourceQuotaStatus implementation.

### Calculate the available resources in the Namespace

To calculate the number of available resources in the given Namespace, list the ResourceQuotas and loop through them.
You can get the available number of resources by calculating the difference from **.spec.hard** and **.status.used** in each ResourceQuota.
For each resource limit specified in the ResourceQuotas, you must calculate the available number of resources using the ResourceQuota with the lowest **.spec.hard** value for that resource type.

### Calculate the necessary resources to create the next replica

To calculate how many resources you need to create the next replica, sum up the resource usage of all containers in the replica Pod.
If some resource limit is not specified directly in the replica's template, check if any LimitRange does not specify it.
In this situation, add the proper LimitRange limit to the list of necessary resources.
When the same limit is specified in the LimitRange and in the replica's template, the replica's template limit has priority.

### Check the ReplicaSets

List ReplicaSets in the given Namespace and loop through them. You must check every ReplicaSet which does not reach the number of desired replicas.
To achieve that, implement the logic which calculates if the ReplicaSet has enough resources to create the next replica.
When there is not enough resources to create the next replica and the number of desired replicas is not reached, return a ResourceQuotasStatus with a flag set to `true`.

### Check the StatefulSets

List StatefulSets in the given Namespace and loop through them. You must check every StatefulSet which does not reach the number of desired replicas.
To achieve that, implement the logic which calculates if the StatefulSet has enough resources to create the next replica.
If there is not enough resources to create the next replica and the number of replicas is not reached, return a ResourceQuotasStatus with a flag set to `true`.

### End of checking

When both checks have passed and the resources usage in the given Namespace did not exceed any ResourceQuota limit, return the ResourceQuotasStatus with a flag set to `false`.

## Examples of the query and the response

The ResourceQuotasStatus query looks as follows:
```graphql
query{
  resourceQuotasStatus(namespace:"production"){
    exceeded
    exceededQuotas{
      quotaName
      resourceName
      affectedResources
    }
  }
}
```
This query returns two types of response: exceeded and not exceeded.
Not exceeded response looks as follows:
```graphql endpoint doc
    "resourceQuotasStatus": {
      "exceeded": false,
      "exceededQuotas": []
    }
```
Exceeded response looks like this:
```graphql endpoint doc
    "resourceQuotasStatus": {
      "exceeded": true,
      "exceededQuotas": [
        {
          "quotaName": "kyma-default",
          "resourceName": "limits.memory",
          "affectedResources": [
            "ReplicaSet/redis-client-5df45c4998",
            "ReplicaSet/redis-client-5df9544d7f"
          ]
        },
        {
          "quotaName": "custom-quota",
          "resourceName": "requests.memory",
          "affectedResources": [
            "StatefulSet/example",
            "ReplicaSet/redis-client-5df9544d7f"
          ]
        },
      ]
    }
```
- **exceeded** equals `true` if any ResourceQuota is exceeded.
- **exceededQuotas** contains the list of exceeded ResourceQuotas limits and set of resources which exceed the limits. The list is empty if there are no exceeded ResourceQuotas in your Namespace.
- **quotaName** represents the name of the ResourceQuota with exceeded limit.
- **resourceName** represents the name of the resource which exceeded the ResourceQuota limit.
- **affectedResources** contains the list of resources which exceed the defined **resourceName** limit from the **quotaName** ResourceQuota.
