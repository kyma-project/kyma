---
title: Deploy with a private Docker registry
type: Details
---

Docker is a free tool to deploy applications and servers. To run an application on Kyma, provide the application binary file as a Docker image located in a Docker registry. Use the `DockerHub` public registry to upload your Docker images for free access to the public. Use a private Docker registry to ensure privacy, increased security, and better availability.

This document shows how to deploy a Docker image from your private Docker registry to the Kyma cluster.

## Details

The deployment to Kyma from a private registry differs from the deployment from a public registry. You must provide Secrets accessible in Kyma, and referenced in the `.yaml` deployment file. This section describes how to deploy an image from a private Docker registry to Kyma. Follow the deployment steps:

1. Create a Secret resource.
2. Write your deployment file.
3. Submit the file to the Kyma cluster.

### Create a Secret for your private registry

A Secret resource passes your Docker registry credentials to the Kyma cluster in an encrypted form. For more information on Secrets, refer to the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/secret/).

To create a Secret resource for your Docker registry, run the following command:

```bash
kubectl create secret docker-registry {secret-name} --docker-server={registry FQN} --docker-username={user-name} --docker-password={password} --docker-email={registry-email} --namespace={namespace}  
```

Refer to the following example:

```bash
kubectl create secret docker-registry docker-registry-secret --docker-server=myregistry:5000 --docker-username=root --docker-password=password --docker-email=example@github.com --namespace=production
```

The Secret is associated with a specific Namespace. In the example, the Namespace is `production`. However, you can modify the Secret to point to any desired Namespace.

### Write your deployment file

1. Create the deployment file with the `.yml` extension and name it `deployment.yml`.

2. Describe your deployment in the `.yml` file. Refer to the following example:

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     namespace: production # {production/stage/qa}
     name: my-deployment # Specify the deployment name.
     annotations:
       sidecar.istio.io/inject: true
   spec:
     replicas: 3 # Specify your replica - how many instances you want from that deployment.
     selector:
       matchLabels:
         app: app-name # Specify the app label. It is optional but it is a good practice.
     template:
       metadata:
         labels:
           app: app-name # Specify app label. It is optional but it is a good practice.
           version: v1 # Specify your version.
       spec:
         containers:
         - name: container-name # Specify a meaningful container name.
           image: myregistry:5000/user-name/image-name:latest # Specify your image {registry FQN/your-username/your-space/image-name:image-version}.
           ports:
             - containerPort: 80 # Specify the port to your image.
         imagePullSecrets:
           - name: docker-registry-secret # Specify the same Secret name you generated in the previous step for this Namespace.
           - name: example-secret-name # Specify your Namespace Secret, named `example-secret-name`.
   ```

3. Submit your deployment file using this command:

   ```bash
   kubectl apply -f deployment.yml
   ```

Your deployment is now running on the Kyma cluster.
