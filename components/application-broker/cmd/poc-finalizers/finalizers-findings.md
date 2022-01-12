## Finalizers

### Overview

Finalizers concept is described in the [k8s finalizers page](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/#finalizers). This pull branch contains a separate cmd, which implements protection controller. Such controller plays with finalizers
Finalizers allow controllers to implement asynchronous pre-delete hooks. Custom objects support finalizers just like built-in objects.

### Description

The protection controller adds a finalizer to Application, which checks, if any ApplicationMapping exists. It there is no ApplicationMapping with the same name - finalizer is removed and object is deleted.

The protection controller implementation is similar to the [k8s PVC protection controller](https://github.com/kubernetes/kubernetes/blob/f4472b1a92877ed4b1576e7e44496b0de7a8efe2/pkg/controller/volume/pvcprotection/pvc_protection_controller.go)

### Building and running

Build:
```bash
go build cmd/poc-finalizers/main.go cmd/poc-finalizers/controller.go
```

Run:
```bash
./main
```

### How the controller works

The protection controller listens on Application events. In case of an update, checks if must be performed any change on the object. If the object is just created - the controller adds the finalizer. If the object is being deleted (`metadata.deletionTimestamp` is set) - the controller checks, if the object can be deleted. If yes, it removes the finalizer. The k8s resource update is safe because an optimistick lock failure will occur. It can be handled by `IsConflict` from `k8s.io/apimachinery/pkg/api/errors` package. You can find more about optimistic lock in [Resource Operation section k8s api doc](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resource-operations). 

### Demo

1. Run the controller:
Run:
```bash
> ./main
```

2. Create application resource
```bash
> kubectl apply -f cmd/poc-finalizers/application-prod.yaml
```

3. Check, if finalizers were applied
```bash
> kubectl get app ec-prod -o jsonpath='{.metadata.finalizers}'; echo
[protection-finalizer]
```

4. Create ApplicationMapping, which blocks Application deletion
```bash
> kubectl apply -f cmd/poc-finalizers/mapping-prod.yaml
```

5. Try to remove Application
```bash
> kubectl delete app ec-prod
application "ec-prod" deleted
```

6. Application still exists, let's check the metadata of the object:
```bash
> kubectl describe app ec-prod
Metadata:
  Cluster Name:                   
  Creation Timestamp:             2018-03-19T10:11:59Z
  Deletion Grace Period Seconds:  0
  Deletion Timestamp:             2018-03-19T10:26:09Z
  Finalizers:
    protection-finalizer
  Generation:        0
  Resource Version:  13479
  Self Link:         /apis/application.kyma.io/v1alpha1/ec-prod
  UID:               f5eaea15-2b5d-11e8-9892-080027ab8e2d
```

7. The finalizers was not removed because ApplicationMapping still exists. Let's remove ApplicationMapping:
```bash
> kubectl delete em ec-prod
```

8. and do update on Application (the controller is not watching ApplicationMapping, which must be done in the production code) to trigger the controller:
```bash
> kubectl delete em ec-prod
applicationmapping "ec-prod" deleted

> kubectl label app ec-prod my-label=awesome
application "ec-prod" labeled
```

9. List Applications:
```bash
> kubectl get app
No resources found.
```
