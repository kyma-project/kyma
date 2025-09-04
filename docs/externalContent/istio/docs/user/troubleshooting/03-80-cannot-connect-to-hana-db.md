# SAP HANA Database Connection Issues

## Symptom

You're unable to connect an application to a SAP HANA Database instance.

## Cause
The Istio module's default configuration does not restrict outbound traffic. This means that the application should have no issues connecting to a SAP HANA Database instance. If you are experiencing issues, they may be related to the SAP HANA Database instance or your cluster configuration. To identify the source of the problem, follow the troubleshooting steps.

## Solution

### Connect to the SAP HANA Database Instance from Outside of the Cluster
1. Download SAP HANA Client for your operating system from [SAP Development Tools](https://tools.hana.ondemand.com/#hanatools).
2. Unpack the downloaded archive.
3. Install SAP HANA Client.
4. To connect to SAP HANA Database instance, use the following command:
    ```bash
    hdbsql -n {HANA_DB_INSTANCE_ADDRESS} -u {HANA_DB_USER} -p {HANA_DB_PASSWORD}
    ```
    For example:
    ```bash
    hdbsql -n aaa.bbb.ccc.ddd:30015 -u my_user -p mypassword
    ```
   If the connection is successful and you can execute queries, the issue is not related to the SAP HANA Database instance.

### Connect to the SAP HANA Database Instance from Inside of the Cluster
1. Build a Docker image with the SAP HANA Client installed. You can use the following Dockerfile:
    ```Dockerfile
    FROM eclipse-temurin:17
    WORKDIR /build
    COPY client.tar client.tar
    RUN tar -xvf client.tar
    RUN echo "/usr/local/bin" | ./client/hdbinst

    ENTRYPOINT ["sleep", "8000"]
    ```
2. Download the SAP HANA Client for Linux x86 64-bit from [SAP Development Tools](https://tools.hana.ondemand.com/#hanatools) and save it as `client.tar` in the same directory as the Dockerfile. 
3. To build the image, run the following command:
    ```bash
    docker buildx build --platform=linux/amd64 -t hdbsql .
    ```
4. To test your image, run the following command:
    ```bash
    docker run --entrypoint "hdbsql" hdbsql -v
    ```
    You get an output similar to this example:
    ```
    HDBSQL version 2.20.20.1712178305, the SAP HANA Database interactive terminal.
    Copyright 2000-2024 by SAP SE.
    ```
5. Publish the image to a container registry.
6. Run the image in the Kubernetes cluster:
    ```bash
    kubectl create deployment hdbsql --image={PUBLISHED_IMAGE_NAME}
    ```
7. To attach to the Pod and connect to the SAP HANA Database instance, run the following command:
    ```bash
    hdbsql -n {HANA_DB_INSTANCE_ADDRESS} -u {HANA_DB_USER} -p {HANA_DB_PASSWORD}
    ```
   If the connection is successful and you can execute queries, the issue is not related to the setup of the cluster.
8. Check the connection from a Pod that has the Istio sidecar injected. To do this, create a Deployment in a namespace with Istio sidecar injection enabled.