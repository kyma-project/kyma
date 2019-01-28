---
title: CLI reference
---

 Management of the Event Bus is based on the custom resources specifically defined for Kyma. Manage all of these resources through [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/).

## Details

This section describes the resource names to use in the kubectl command line, the command syntax, and examples of use.

### Resource types

Event Bus operations use the following resources:

| Singular name        | Plural name         |
| -------------------- | ------------------- |
| subscription         | subscriptions       |

### Syntax

Follow the `kubectl` syntax, `kubectl {command} {type} {name} {flags}`, where:

* {command} is any command, such as `describe`.
* {type} is a resource type, such as `clusterserviceclass`.
* {name} is the name of a given resource type. Use {name} to make the command return the details of a given resource.
* {flags} specifies the scope of the information. For example, use flags to define the Namespace from which to get the information.

### Examples

The following examples show how to create new Subscriptions, list them, and obtain detailed information on their statuses.

* Create a new Subscription directly from the terminal:

```
   cat <<EOF | kubectl create -f -
   apiVersion: eventing.kyma-project.io/v1alpha1
   kind: Subscription
   metadata:
     name: my-subscription
     namespace: stage
   spec:
     endpoint: http://testjs.default:8080/
     push_request_timeout_ms: 2000
     max_inflight: 400
     include_subscription_name_header: true
     event_type: order_created
     event_type_version: v1
     source_id: stage.commerce.kyma.local
EOF
```

* Get the list of all Subscriptions:

```
kubectl get subscription --all-namespaces
```

* Get the list of all Subscriptions with detailed information on the Subscription status:

```
kubectl get subscriptions -n stage -o=custom-columns=NAME:.metadata.name,STATUS:.status.conditions[*].status,STATUS\ TYPE:.status.conditions[*].type
```
