---
title: Deploy a new Remote Environment
type: Getting Started
---

The Remote Environment Controller provisions and de-provisions necessary deployments for the created Remote Environments.

The following operations are available:

- Create a new Remote Environment
- Delete the Remote Environment
- Update the Remote Environment configuration

You can perform all these operations using the Console UI or kubectl.


All Remote Environments are installed in the `kyma-integration` Namespace.

>**NOTE:** A Remote Environment represents a single connected external solution.


## Install a Remote Environment

You can install a Remote Environment using either the Console UI or kubectl.

### Using Console:

- Go to the Kyma Console.
- Select **Administration**.
- Select the **Remote Environments** from the **Integration** section.
- Click **Create Remote Environment**..

![Add RE](./assets/create-re.png)

- Provide the following details:
    - Name
    - Description
    - Optional labels of your choice which are key-value pairs

![Update RE](./assets/edit-re.png)

 - Click **Create**.

The new Remote Environment is created. You can check its status in the **Remote Environment** view.

### Using kubectl

You can create a new Remote Environment using the kubectl `apply` command for the `re-production-1.yaml` file:

re-production-1.yaml:

``` yaml
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

Run the following command:

``` bash
kubectl apply -f ./re-production-1.yaml
```

### How to check if your remote environment was successfully created.

The new Remote Environment appears on the **Remote Environments** list with the `Serving` status.

## Delete a Remote Environment

You can remove a Remote Environment from Kyma using either the Console UI or kubectl.


### Using Console:

- Go to the Kyma console UI.
- Select **Administration**.
- Select the **Remote Environments** from the **Integration** section.
- Choose the Remote Environment you want to delete.
- Click **Delete**.

![Delete RE](./assets/delete-re.png)


### Using kubectl

Delete the Remote Environment using the following command:

```bash
kubectl delete re name-of-remote-environment
```

## Update a Remote Environment

You can update a Remote Environment using either the Console UI or kubectl.

>**NOTE:** You cannot change the name of a Remote Environment.

### Using Console:

- Go to the Kyma Console UI.
- Select **Administration**.
- Select the **Remote Environments** from the **Integration** section.
- Choose the Remote Environment to which you want to update.
- Change the description and labels.
- Click **Save**.

### Using kubectl

Update the `re-production-1.yaml` file

``` yaml
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

Run the following command:

``` bash
kubectl apply -f ./re-production-1.yaml
```