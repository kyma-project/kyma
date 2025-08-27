# Serverless Buildless Mode

Learn how to accelerate your development with Serverless buildless mode.

From the beginning, Kyma Serverless has aimed to accelerate the development of fast prototypes by allowing users to focus on business logic rather than containerization and Kubernetes deployment. Our goal is to remove operational barriers so developers can iterate quickly and efficiently.

With the introduction of buildless mode, we significantly shortened the feedback loop during prototype development by eliminating the image build step in Kyma runtime. In buildless mode, instead of building and pushing custom Function images into the in-cluster registry, your code and dependencies are mounted into Kyma-provided runtime images. This approach positions Kyma Serverless as a more efficient development tool, enabling even faster iteration. Additionally, it eliminates the architectural complexities and limitations of deploying Serverless Functions on Kubernetes.

## Features

- **Faster deployment**: Even though Function dependencies are resolved and downloaded at the start time of each Function, the overall time required for the Function to become ready is significantly shorter, as there is no need to wait for a build Job to complete before the Function Pod is scheduled.
- **Resource efficiency**: Eliminates the need for Serverless to acquire computational resources from worker nodes to build the image.
- **Enhanced security**: By eliminating build jobs, Functions can run in namespaces with more restrictive Pod security levels enabled.
- **No additional storage required**: No additional storage resources are used to store the Function image.
- **Simplified Architecture**: The Serverless module no longer requires Docker Registry, making it more lightweight and easier to manage.

## Outcomes of Switching to Buildless Mode

- The internal resources used for storing custom Function images (Docker Registry) are uninstalled from the Serverless module
- Your existing Functions are redeployed without downtime and started as Pods based on Kyma-provided images with your handler code and dependencies mounted.
- Build Jobs associated with your Function are deleted.
- All fields that were deprecated in the Serverless [1.6.0 release](https://github.com/kyma-project/serverless/releases/tag/1.6.0) are no longer functional in buildless mode:
  - [Function CRD](https://kyma-project.io/#/serverless-manager/user/resources/06-10-function-cr)
    - `spec.scaleConfig` - for existing Functions with `scaleConfig` defined, the `HorizontalPodAutoscaler` objects are not deleted upon switching to buildless mode, but are no longer managed by the Serverless module. To learn how to scale Functions, see [Use External Scalers](https://kyma-project.io/#/serverless-manager/user/tutorials/01-130-use-external-scalers).
    - `spec.resourceConfiguration.​build`
  - [Serverless CRD](https://kyma-project.io/#/serverless-manager/user/resources/06-20-serverless-cr)
    - `spec.dockerRegistry`
    - `spec.targetCPUUtilizationPercentage`
    - `spec.functionBuildExecutorArgs`
    - `spec.functionBuildMaxSimultaneousJobs`
    - `spec.defaultBuildJobPreset`

## Using Fixed Dependency Versions

- **Avoid using `latest` versions of Function dependencies**: Since dependencies are resolved at Function's Pod start time in buildless mode, using `latest` versions can lead to inconsistencies between replicas of the same Function. This may be the case when the dependency provider releases a new version after one replica is already running and before another replica is created due to auto-scaling.  Always specify exact versions of dependencies to ensure stability and predictability.
- **Dependency resolution behavior**: Be aware that each replica of a Function may resolve and use a different dependency version if the version is not explicitly pinned.

## Switching to Buildless Serverless

To enable buildless mode for Serverless, you must add the relevant annotation in the Serverless custom resource (CR). Follow these steps:

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to Kyma dashboard.

2. Choose **Modify Modules**, and in the **View** tab, choose `serverless`.

3. Go to **Edit**.

4. In **Annotations**, enter `serverless.kyma-project.io/buildless-mode` as the key, and `enabled` as the value. Save the changes.

You have enabled Serverless buildless mode.
</Tab>
<Tab name="kubectl">

1. Edit the Serverless CR:

   ```bash
   kubectl edit -n kyma-system serverlesses.operator.kyma-project.io default
   ```

2. In the `metadata` section of the CR, add the following annotation:

   ```yaml
   annotations:
   serverless.kyma-project.io/buildless-mode: "enabled"
   ```

3. Save the file to apply the changes.

You have enabled Serverless buildless mode.
</Tab>
</Tabs>

To disable buildless mode for Serverless, remove the annotation.
