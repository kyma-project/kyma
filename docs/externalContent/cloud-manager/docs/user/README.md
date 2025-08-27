
# Cloud Manager Module

Use the Cloud Manager module to manage infrastructure providers' resources from SAP BTP, Kyma runtime.

## What is Cloud Manager?

The Cloud Manager module manages access to cloud providers' chosen resources and products. Once you add Cloud Manager to your Kyma cluster, the module securely brings the offered resources. Cloud Manager is strictly coupled with the cloud provider where your Kyma cluster is deployed.

> [!NOTE]
> Using certain features of the Cloud Manager module introduces additional costs. For more information, see [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#loioc33bb114a86e474a95db29cfd53f15e6__section_cloud_manager).

## Scope

The Cloud Manager module is available on Kyma clusters running on Amazon Web Services, Google Cloud, and Microsoft Azure.

> [!WARNING]
> Cloud Manager supports the NFS and VPC peering feature of Amazon Web Services and Google Cloud only.
> However, SAP-internal users can also benefit from the Cloud Manager modules' support for the VPC peering feature of Microsoft Azure. <!-- VPC peering for Microsoft Azure is visible only in the Internal DRAFT version of Help Portal docs and it is not part of the Cloud Production version of Help Portal docs -->

## Features

The Cloud Manager module provides the following features tailored for each of the cloud providers.

* [NFS](./00-20-nfs.md): Network File System (NFS) server that can be used as a ReadWriteMany (RWX) volume in the Kyma cluster.
* [VPC peering](./00-30-vpc-peering.md): Virtual Private Cloud (VPC) peering between your SAP BTP, Kyma runtime and remote cloud provider's project, account, or subscription.
* [Redis](./00-40-redis.md): cloud provider-flavored cache that can be used in your Kyma cluster.
* [Redis cluster](./00-50-redis-cluster.md): cloud provider-flavored cache in cluster mode that can be used in your Kyma cluster.

## Architecture

Cloud Manager has read and write access to your IpRange, VpcPeering, NfsVolume, and Redis custom resources in the Kyma cluster. The module also manages Kyma VPC networks, NFS Volume instances, and Redis instances in your cloud provider subscription in Kyma.

![Cloud Manager Architecture](./assets/cloud-manager-architecture.drawio.svg)

## API / Custom Resources Definitions

The `cloud-resources.kyma-project.io` Custom Resource Definition (CRD) describes the data kind and format that Cloud Manager uses to configure resources. For more information, see [Cloud Manager Resources](./resources/README.md) (CRs).

## Related Information

* [Cloud Manager module tutorials](./tutorials/README.md)
* [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#calculation-with-the-cloud-manager-module)
