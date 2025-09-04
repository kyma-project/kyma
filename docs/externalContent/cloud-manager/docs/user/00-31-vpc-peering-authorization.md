# Authorizing Cloud Manager in the Remote Cloud Provider

Learn about the entitlements for the Cloud Manager module in the remote cloud provider required to use the VPC peering feature.

> [!WARNING]
> Cloud Manager supports the VPC peering feature of Amazon Web Services and Google Cloud only.
> However, SAP-internal users can also benefit from the Cloud Manager modules' support for the VPC peering feature of Microsoft Azure.

To create VPC peering in SAP BTP, Kyma runtime, you must authorize the Cloud Manager module in the remote cloud provider to accept the connection.

## Amazon Web Services

For cross-account access in Amazon Web Services, Cloud Manager uses `AssumeRole`. `AssumeRole` requires specifying the trusted principle. For more information, see the [official Amazon Web Services documentation](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/sts/assume-role.html).

Use the following table to identify the Cloud Manager principal. Then, perform the required actions.

| BTP Cockpit URL                    | Kyma Dashboard URL                     | Cloud Manager Principal                                      |
|------------------------------------|----------------------------------------|--------------------------------------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | `arn:aws:iam::194230256199:user/cloud-manager-peering-stage` |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | `arn:aws:iam::194230256199:user/cloud-manager-peering-prod`  |
<!-- The stage landscape is visible only in the Internal DRAFT version of Help Portal docs. The stage landscape is not part of the Cloud Production version of Help Portal docs -->

1. Create a new role named **CloudManagerPeeringRole** with a trust policy that allows the Cloud Manager principal to assume the role:

    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "AWS": "{CLOUD_MANAGER_PRINCIPAL}"
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }

    ```

2. Create a new **CloudManagerPeeringAccess** managed policy with the following permissions:

    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "Statement1",
                "Effect": "Allow",
                "Action": [
                    "ec2:AcceptVpcPeeringConnection",
                    "ec2:DescribeVpcs",
                    "ec2:DescribeVpcPeeringConnections",
                    "ec2:DescribeRouteTables",
                    "ec2:CreateRoute",
                    "ec2:CreateTags"
                ],
                "Resource": "*"
            }
        ]
    }
    ```

3. Attach the **CloudManagerPeeringAccess** policy to the **CloudManagerPeeringRole**.

## Google Cloud

Grant the following permissions to the Kyma service account in your GCP project:

| Permission                           | Description                                                                 |
|--------------------------------------|-----------------------------------------------------------------------------|
| `compute.networks.addPeering`        | Required to create the peering request in the remote project and VPC.       |
| `compute.networks.get`               | Required to fetch the list of existing VPC peerings from the remote VPC.    |
| `compute.networks.listEffectiveTags` | Required to check if the remote VPC is tagged with the Kyma shoot name tag. |

It is recommended to create an Identity and Access Management (IAM) custom role with the permissions listed above. 
For more information, see the official Google Cloud documentation on how to [create a custom role](https://cloud.google.com/iam/docs/creating-custom-roles#creating) or the [Authorize Cloud Manager in the Remote Project section](tutorials/01-30-20-gcp-vpc-peering.md#authorize-cloud-manager-in-the-remote-project) in the Creating VPC Peering in Google Cloud tutorial.

### Service Account

Use the following table to identify the correct Cloud Manager service account:

| BTP Cockpit URL                    | Kyma Dashboard URL                     | Cloud Manager Service Account                                          |
|------------------------------------|----------------------------------------|------------------------------------------------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | `cloud-manager-peering@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com` |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | `cloud-manager-peering@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com`  |
<!-- The stage landscape is visible only in the Internal DRAFT version of Help Portal docs. The stage landscape is not part of the Cloud Production version of Help Portal docs -->

With the service account, you can authorize the Cloud Manager module in the remote project.
For more information, see the official Google Cloud documentation on how to [grant a single role](https://cloud.google.com/iam/docs/granting-changing-revoking-access#grant-single-role) to a service account or the [Authorize Cloud Manager in the Remote Project section](tutorials/01-30-20-gcp-vpc-peering.md#authorize-cloud-manager-in-the-remote-project) in the Creating VPC Peering in Google Cloud tutorial.

## Microsoft Azure
<!-- VPC peering for Microsoft Azure is visible only in the Internal DRAFT version of Help Portal docs and it is not part of the Cloud Production version of Help Portal docs -->

To authorize Cloud Manager in the remote subscription, Microsoft Azure requires specifying the service principal. Use the following table to identify the Cloud Manager service principal. Then, perform the required actions.

| BTP Cockpit URL                    | Kyma Dashboard URL                     | Cloud Manager Service Principal  | Cloud Manager Application (Client) ID |
|------------------------------------|----------------------------------------|----------------------------------|---------------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | kyma-cloud-manager-peering-stage | 8e08320c-7e81-42bd-9eee-e5dae04cadf0  |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | kyma-cloud-manager-peering-prod  | 202aa655-369d-4fe7-bbbc-d033d96a687e  |

1. Verify if the Cloud Manager service principal exists in your tenant.
2. **Optional:** If the service principal doesn't exist, create one for the Cloud Manager application in your tenant.
3. Assign the `Classic Network Contributor` and `Network Contributor` roles to the Cloud Manager service principal.

For more information, see the official Microsoft Azure documentation on how to [Assign Azure roles using the Azure portal](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-portal) and how to [Manage service principals](https://learn.microsoft.com/en-us/azure/databricks/admin/users-groups/service-principals) or the [Authorize Cloud Manager in the Remote Subscription](tutorials/01-30-30-azure-vpc-peering.md#authorize-cloud-manager-in-the-remote-subscription) section in the Creating VPC Peering in Microsoft Azure tutorial.