# Use Docker Registry Internally

This tutorial shows how you can push an image to the Docker Registry and use it.

## Prerequsities

* [Kyma CLI v3](https://github.com/kyma-project/cli)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Docker](https://www.docker.com/)

## Steps

1. Create a `simple.py` file, and paste the following content:

   ```python
   import time
   import datetime
    
   while True:
     print("Current time:", datetime.datetime.now(), flush=True)
     time.sleep(10)
   ```

2. Create `Dockerfile` with the following content:

   ```dockerfile
   FROM python:3-alpine

   WORKDIR /usr/src/app

   COPY simple.py .
   
   CMD [ "python", "./simple.py" ]
   ```

3. Build and push the image to Docker Registry:

   ```bash
   docker build -t simple-image .
   ```

4. Import the image to Docker Registry:

   ```bash
   kyma alpha registry image-import simple-image:latest
   ```

4. Create a Pod using the image from Docker Registry:

   ```bash
   kubectl run simple-pod --image=localhost:32137/simple-image:latest --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'
   ```

5. Check if the Pod is running:

   ```bash
   kubectl get pods simple-pod
   ```

    Expected output:

   ```bash
   NAME         READY   STATUS    RESTARTS   AGE
   simple-pod   1/1     Running   0          60s
   ```

6. Use the following command to print the current time every 10 seconds:

   ```bash
   kubectl logs simple-pod
   ```

   Expected output:

   ```bash
   Current time: 2024-05-17 11:23:34.957406
   Current time: 2024-05-17 11:23:44.954583
   Current time: 2024-05-17 11:23:54.956107
   Current time: 2024-05-17 11:24:04.966306
   ```
