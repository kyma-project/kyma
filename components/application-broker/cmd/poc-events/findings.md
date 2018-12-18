# POC: CRs status changing and event sending

## Run
Example can be executed outside of the cluster:
```
go run cmd/poc-events/events.go
```
## Results

### Events
See `events.go` which follow example from:
`https://kubernetes.io/blog/2018/01/reporting-errors-using-kubernetes-events/`.

Beware, that events by default, have **retention set to 1 hour**, see [kube-apiserver configuration options](https://kubernetes.io/docs/reference/generated/kube-apiserver/):
> --event-ttl  duration   Amount of time to retain events. (default 1h0m0s)

### Status changing
#### TLDR
Currently changing status of custom resource does not work. Probably will be released in k8s v1.10

#### Detailed description
Currently changing status of custom resource does not work, but [proposal](https://github.com/kubernetes/community/pull/913) is already accepted.
Implementation is also merged to master and tagged with version: "v1.10.0-beta.4". See [Add subresources for custom resources](https://github.com/kubernetes/kubernetes/pull/55168/commits/6fbe8157e39f6bd7ad885a8a6f8614e2fe45dc5e)

So currently we get following error when calling UpdateStatus method.
```text
panic: while updating status: the server could not find the requested resource (put remoteenvironments.remoteenvironment.kyma.io ec-prod)
```

To add `UpdateStatus` method to Remote Environment client, following actions has to be performed:
- add Status field for type `RemoteEnvironment`
- remove annotation `+genclient:noStatus`

[API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status) contains information how Status should looks like:

>Spec and Status:

>By convention, the Kubernetes API makes a distinction between the specification of the desired state of an object (a nested object field called "spec") and the status of the object at the current time (a nested object field called "status").
>The PUT and POST verbs on objects MUST ignore the "status" values, to avoid accidentally overwriting the status in read-modify-write scenarios. A /status subresource MUST be provided to enable system components to update statuses of resources they manage.

>Conditions represent the latest available observations of an object's current state. Objects may report multiple conditions, and new types of conditions may be added in the future. Therefore, conditions are represented using a list/slice, where all have similar structure.

>The FooCondition type for some resource type Foo must contain at least type and status fields.

I found [implementation](https://github.com/jetstack/cert-manager/blob/master/pkg/apis/certmanager/v1alpha1/types.go) which follows those conventions about Status field. Currently they are using `Update`, instead `UpdateStatus` which seems to be against conventions.
