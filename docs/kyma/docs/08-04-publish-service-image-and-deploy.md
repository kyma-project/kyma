---
title: Publish a service Docker image and deploy it to Kyma
type: Tutorials
---

Follow this tutorial to learn how to develop a service locally. You can immediately see all the changes made in a local Kyma installation based on Minikube, without building a Docker image and publishing it to a Docker registry, such as the Docker Hub.

Using the same example service, this tutorial explains how to build a Docker image for your service, publish it to the Docker registry, and deploy it to the local Kyma installation. The instructions base on Minikube, but you can also use the image that you create and the Kubernetes resource definitions that you use on the Kyma cluster.

>**NOTE:** The deployment works both on local Kyma installation and on the Kyma cluster.

## Steps

### Build a Docker image

The `http-db-service` example used in this guide provides you with the `Dockerfile` necessary for building Docker images. Examine the `Dockerfile` to learn how it looks and how it uses the Docker Multistaging feature, but do not use it one-to-one for production. There might be custom `LABEL` attributes with values to override.

1. In your terminal, go to the `examples/http-db-service` directory. If you did not follow the [Sample service deployment on local](#tutorials-sample-service-deployment-on-local) guide and you do not have this directory locally, get the `http-db-service` example from the [`examples`](https://github.com/kyma-project/examples) repository.
2. Run the build with `./build.sh`.

>**NOTE:** Ensure that the new image builds and is available in your local Docker registry by calling `docker images`. Find an image called `example-http-db-service` and tagged as `latest`.

### Register the image in the Docker Hub

This guide bases on Docker Hub. However, there are many other Docker registries available. You can use a private Docker registry, but it must be available in the Internet. For more details about using a private Docker registry, see the [this](#tutorials-publish-a-service-docker-image-and-deploy-it-to-kyma) document.

1. Open the [Docker Hub](https://hub.docker.com/) webpage.
2. Provide all of the required details and sign up.

### Sign in to the Docker Hub registry in the terminal

1. Call `docker login`.
2. Provide the username and password, and select the `ENTER` key.

### Push the image to the Docker Hub

1. Tag the local image with a proper name required in the registry: `docker tag example-http-db-service {USERNAME}/example-http-db-service:0.0.1`.
2. Push the image to the registry: `docker push {USERNAME}/example-http-db-service:0.0.1`.

>**NOTE:** To verify if the image is successfully published, check if it is available online at the following address: `https://hub.docker.com/r/{USERNAME}/example-http-db-service/`

### Deploy to Kyma

The `http-db-service` example contains sample Kubernetes resource definitions needed for the basic Kyma deployment. Find them in the `deployment` folder. Perform the following modifications to use your newly-published image in the local Kyma installation:

1. Go to the `deployment` directory.
2. Edit the `deployment.yaml` file. Change the **image** attribute to `{USERNAME}/example-http-db-service:0.0.1`.
3. Create the new resources in local Kyma using these commands: `kubectl create -f deployment.yaml -n stage && kubectl create -f ingress.yaml -n stage`.
4. Edit your `/etc/hosts` to add the new `http-db-service.kyma.local` host to the list of hosts associated with your `minikube ip`. Follow these steps:
    - Open a terminal window and run: `sudo vim /etc/hosts`
    - Select the **i** key to insert a new line at the top of the file.
    - Add this line: `{YOUR.MINIKUBE.IP} http-db-service.kyma.local`
    - Type `:wq` and select the **Enter** key to save the changes.
5. Run this command to check if you can access the service: `curl https://http-db-service.kyma.local/orders`. The response should return an empty array.
