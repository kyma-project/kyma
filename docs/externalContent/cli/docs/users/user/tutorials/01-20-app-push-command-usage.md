# Running an Application Using the `app push` Command

This tutorial shows how you can deploy your application using Kyma CLI.

## Deploy Your Application From the Source Code

To use `kyma alpha app push`, you must also provide either a Dockerfile, Docker image or the application's source code. In this tutorial, we use the application source code. For example, you can use one of the code samples from the [Paketo Buildpacks code examples](https://github.com/paketo-buildpacks/samples/tree/main), but in this tutorial we work on the [Paketo Buildpacks java maven app](https://github.com/paketo-buildpacks/samples/tree/main/java/maven).

1. Add the Istio, Docker Registry, and API Gateway modules:

   ```bash
   kyma alpha module add istio --default-cr
   kyma alpha module add docker-registry -c experimental --default-cr
   kyma alpha module add api-gateway --default-cr
   ```

2. Clone the [Paketo code examples](https://github.com/paketo-buildpacks/samples/tree/main) repository into the desired folder:

   ```url
   git clone https://github.com/paketo-buildpacks/samples.git
   ```

3. Navigate to the `java/maven` directory:

   ```bash
   cd java/maven
   ```

4. Deploy your application.

   > [!NOTE]
   > Besides the required `--name` flag, you must also use the `--code-path` flag to run the application from the source code.

   To deploy your application on a cluster with its own APIRule allowing outside access, run the following command in the current directory:

   ```bash
   kyma alpha app push --name=Test-App --code-path=. --container-port=8888 --expose
   ```

   > [!NOTE]
   > Depending on your needs, you can also create deployments of your applications without `--expose` or `--container-port` flags. This changes the way you communicate with your application.

5. Copy the URL address you should get after deploying your application. You will use it in the next step.

6. Check the deployed application connection

   To check if the deployed application connection is working properly, run the following curl request:

   ```bash
   curl {ADDRESS-RETURNED-FROM-THE-APP-PUSH}:8080/actuator/health
   ```  

   You should get the following response:

   ```json
   {"status":"UP","groups":["liveness","readiness"]}
   ```

   > [!NOTE]
   > Depending on how you deploy your application, the way you communicate with it differs.
   >
   > Without `--container-port`, the `app push` command doesn't create a Service resource, and you must port-forward the deployment in one terminal and then check the health of an application in another:
   >
   > kubectl port-forward deployment/Test-App 8080:8080
   >
   > curl localhost:8080/actuator/health
   >
   > Without `--expose`, the `app push` command doesn't create an APIRule resource, and you must port-forward the service in one terminal window and then check the health of an application in another because your application is not be exposed to the outside of the cluster:
   >
   > kubectl port-forward svc/Test-App 8080:8080
   >
   > curl localhost:8080/actuator/health
