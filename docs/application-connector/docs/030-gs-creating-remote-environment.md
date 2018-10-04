---
title: Create a new Remote Environment
type: Getting Started
---

The Remote Environment Controller provisions and de-provisions the necessary deployments for the created Remote Environment (RE).

The following operations are available:
- Create a new Remote Environment
- Delete a Remote Environment
- Update the Remote Environment configuration

You can perform all these operations using either the Command Line Interface (CLI) or the Console UI.

The controller creates all Remote Environments in the `kyma-integration` Namespace.

>**NOTE:** A Remote Environment represents a single connected external solution.

## Create a new Remote Environment

To create a new RE, first create a `re-production-1.yaml` following this template:

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: production-1
spec:
  description: This is a Remote Environment for connecting production system 1.
  labels:
    region: us
    kind: production
```

Use the `kubectl apply` command on the `re-production-1.yaml` file. Run:

```
kubectl apply -f ./re-production-1.yaml
```

Alternatively, create a new Remote Environment through the Console UI.  

- Go to the Kyma Console.
- Select **Administration**.
- Select **Remote Environments** from the **Integration** section.
- Click **Create Remote Environment**.
- Fill in the RE name, description, and add any optional values which are key-value pairs.
- Click **Create**.

You can check the status of the created RE in the **Remote Environment** view of the UI. All properly functioning REs appear on the list with the `Serving` status.

## Delete a Remote Environment

To delete a RE, run this command:

```
kubectl delete re {RE_NAME}
```

Alternatively, use the Console UI:

- Go to the Kyma console UI.
- Select **Administration**.
- Select **Remote Environments** from the **Integration** section.
- Choose the RE you want to delete.
- Click **Delete**.


## Update a Remote Environment

You can update a Remote Environment using either the Console UI or kubectl.

>**NOTE:** You cannot change the name of a Remote Environment.

Update the `re-production-1.yaml` you used to create the RE:

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: production-1
spec:
  description: This is a new description.
  labels:
    region: new-region
    kind: production
```

Run this command:
```
kubectl apply -f ./re-production-1.yaml
```

Alternatively, use the Console UI:

- Go to the Kyma Console UI.
- Select **Administration**.
- Select **Remote Environments** from the **Integration** section.
- Choose the RE to which you want to update.
- Change the description and labels.
- Click **Save**.
