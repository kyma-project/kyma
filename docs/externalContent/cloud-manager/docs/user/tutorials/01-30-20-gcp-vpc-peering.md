# Creating VPC Peering in Google Cloud

This tutorial explains how to create a Virtual Private Cloud (VPC) peering connection between a remote VPC network and SAP BTP, Kyma runtime in Google Cloud.

## Prerequisites

* You have the Cloud Manager module added. See [Add and Delete a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/enable-and-disable-kyma-module?state=DRAFT&version=Internal#loio1b548e9ad4744b978b8b595288b0cb5c).
* Google Cloud CLI

> [!NOTE]
> Use a POSIX-compliant shell or adjust the commands accordingly. For example, if you use Windows, replace the `export` commands with `set` and use `%` before and after the environment variables names.

## Steps

### Authorize Cloud Manager in the Remote Project

To create a VPC peering connection, certain permissions are required in the remote project.
The recommended approach is creating a role and assigning it to the Cloud Manager service account.

1. Export your remote project ID and desired role name as an environment variable.

   ```shell
   export YOUR_REMOTE_PROJECT_ID={YOUR_REMOTE_PROJECT_ID}
   export ROLE_NAME=peeringWithKyma
   ```

2. Create a custom role with the required permissions.

   ```shell
   gcloud iam roles create $ROLE_NAME --permissions="compute.networks.addPeering,compute.networks.get,compute.networks.listEffectiveTags" --project=$YOUR_REMOTE_PROJECT_ID
   ```

3. Check on [Authorizing Cloud Manager in the Remote Cloud Provider](../00-31-vpc-peering-authorization.md#service-account) and assign the custom role created on the previous step to the correct Cloud Manager service account for your environment.

   ```shell
    gcloud projects add-iam-policy-binding $YOUR_REMOTE_PROJECT_ID --member=serviceAccount:cloud-manager-peering@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com --role=projects/$YOUR_REMOTE_PROJECT_ID/roles/$ROLE_NAME
   ```

### Allow SAP BTP, Kyma Runtime to Peer with Your Network

Due to security reasons, the VPC network in the remote project, which receives the VPC peering connection, must contain a tag with the Kyma shoot name.

1. Fetch your Kyma ID and export it as an environment variable.

   ```shell
   export KYMA_SHOOT_ID=`kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}'`
   ```

2. Export your project ID and VPC network as environment variables.

    ```shell
     export REMOTE_PROJECT_ID={YOUR_REMOTE_PROJECT_ID}
     export REMOTE_VPC_NETWORK={REMOTE_VPC_NETWORK}
     ```

3. Create a tag key with the Kyma shoot name in the remote project.

   ```shell
   gcloud resource-manager tags keys create $KYMA_SHOOT_ID --parent=projects/$REMOTE_PROJECT_ID
   ```

4. Create a tag value in the remote project.

   ```shell
   export TAG_VALUE=None
   gcloud resource-manager tags values create $TAG_VALUE --parent=$REMOTE_PROJECT_ID/$KYMA_SHOOT_ID
   ```

5. Fetch the network `selfLinkWithId` from the remote VPC network.

    ```shell
    gcloud compute networks describe $REMOTE_VPC_NETWORK
    ```

    The command returns an output similar to this one:

    ```shell
    ...
    routingConfig:
    routingMode: REGIONAL
    selfLink: https://www.googleapis.com/compute/v1/projects/remote-project-id/global/networks/remote-vpc
    selfLinkWithId: https://www.googleapis.com/compute/v1/projects/remote-project-id/global/networks/1234567890123456789
    subnetworks:
    - https://www.googleapis.com/compute/v1/projects/remote-project-id/regions/europe-west12/subnetworks/remote-vpc
    ...
    ```

6. Export resource ID as an environment variable. Use the value of `selfLinkWithId` returned in the previous command's output, but replace `https://www.googleapis.com/compute/v1` with `//compute.googleapis.com`.

    ```shell
    export RESOURCE_ID="//compute.googleapis.com/projects/remote-project-id/global/networks/1234567890123456789"
    ```

7. Add the tag to the VPC network.

    ```shell
    gcloud resource-manager tags bindings create --tag-value=$TAG_VALUE --parent=$RESOURCE_ID
    ```

### Create VPC Peering

1. Create a GcpVpcPeering resource manifest file.

   ```shell
   cat <<EOF > vpc-peering.yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   ckind: GcpVpcPeering
    metadata:
        name: "vpcpeering-dev"
    spec:
        remotePeeringName: "my-project-to-kyma-dev"
        remoteProject: "remote-project-id"
        remoteVpc: "remote-vpc-network"
        importCustomRoutes: false
    EOF
    ```

2. Apply the manifest file.

   ```shell
   kubectl apply -f vpc-peering.yaml
   ```

   This operation usually takes less than 2 minutes. To check the status of the VPC peering, run:

   ```shell
   kubectl get gcpvpcpeering vpcpeering-dev -o yaml
   ```

   The command returns an output similar to this one:

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpVpcPeering
     finalizers:
     - cloud-control.kyma-project.io/deletion-hook
       generation: 2
       name: vpcpeering-dev
       resourceVersion: "12345678"
       uid: 8545cdaa-66d3-4fa7-b20b-7c716148552f
       spec:
       remotePeeringName: my-project-to-kyma-dev
       remoteProject: remote-project-id
       remoteVpc: remote-vpc-network
       status:
       conditions:
       - lastTransitionTime: "2024-08-12T15:29:59Z"
         message: VpcPeering: my-project-to-kyma-dev is provisioned
         reason: Ready
         status: "True"
         type: Ready
   ```

   The **status.conditions** field contains information about the VPC Peering status.

## Next Steps

When the VPC peering is not needed anymore, you can remove it.

1. Delete the GcpVpcPeering resource from your Kyma cluster.

   ```shell
   kubectl delete gcpvpcpeering vpcpeering-dev
   ```

2. Remove the inactive VPC peering from the remote project.

   ```shell
   gcloud compute networks peerings delete my-project-to-kyma-dev --network=remote-vpc-network --project=remote-project-id
   ```
