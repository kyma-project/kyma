# Creating VPC Peering in Microsoft Azure

> [!Warning]
> VPC peering for Microsoft Azure is a feature available only for SAP-internal teams.

This tutorial explains how to create a Virtual Private Cloud (VPC) peering connection between a remote VPC network and SAP, BTP Kyma runtime in Microsoft Azure. Learn how to create a new resource group, VPC network and a virtual machine (VM), and assign required roles to the provided Kyma service principal in your Microsoft Azure subscription.

## Prerequisites

* You have the Cloud Manager module added. See [Add and Delete a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/enable-and-disable-kyma-module?state=DRAFT&version=Internal#loio1b548e9ad4744b978b8b595288b0cb5c).
* Azure CLI

## Steps

### Authorize Cloud Manager in the Remote Subscription

1. Log in to your Microsoft Azure account and set the active subscription:

   ```shell
   export SUBSCRIPTION={SUBSCRIPTION}
   az login
   az account set --subscription $SUBSCRIPTION
   ```

2. Verify if the Cloud Manager service principal exists in your tenant.
   ```shell
   export APPLICATION_ID={APPLICATION_ID}
   az ad sp show --id $APPLICATION_ID
   ```
3. **Optional:** If the service principal doesn't exist, create one for the Cloud Manager application in your tenant.
   ```shell
   az ad sp create --id $APPLICATION_ID
   ```
4. Assign the required `Classic Network Contributor` and `Network Contributor` Identity and Access Management (IAM) roles to the Cloud Manager service principal. See [Authorizing Cloud Manager in the Remote Cloud Provider](../00-31-vpc-peering-authorization.md#microsoft-azure) to identify the Cloud Manager principal.
    ```shell
    export SUBSCRIPTION_ID=$(az account show --query id -o tsv)
    export OBJECT_ID=$(az ad sp show --id $APPLICATION_ID --query "id" -o tsv)
    
    az role assignment create --assignee $OBJECT_ID \
    --role "Network Contributor" \
    --scope "/subscriptions/$SUBSCRIPTION_ID"
   
    az role assignment create --assignee $OBJECT_ID \
    --role "Classic Network Contributor" \
    --scope "/subscriptions/$SUBSCRIPTION_ID"

### Set Up a Test Environment in the Remote Subscription

1. Set the region that is closest to your Kyma cluster. Use `az account list-locations` to list available locations.

   ```shell
   export REGION={REGION}
   ```

2. Create a resource group as a container for related resources:

   ```shell
   export RANDOM_ID="$(openssl rand -hex 3)"
   export RESOURCE_GROUP_NAME="myResourceGroup$RANDOM_ID"
   az group create --name $RESOURCE_GROUP_NAME --location $REGION
   ```

3. Create a network:

   ```shell
   export VNET_NAME="myVnet$RANDOM_ID"
   export ADDRESS_PREFIX=172.0.0.0/16
   export SUBNET_PREFIX=172.0.0.0/24
   export SUBNET_NAME="MySubnet"

   az network vnet create -g $RESOURCE_GROUP_NAME -n $VNET_NAME --address-prefix $ADDRESS_PREFIX --subnet-name $SUBNET_NAME --subnet-prefixes $SUBNET_PREFIX
   ```

4. Create a virtual machine (VM):

   ```shell
   export VM_NAME="myVM$RANDOM_ID"
   export VM_IMAGE="Canonical:0001-com-ubuntu-minimal-jammy:minimal-22_04-lts-gen2:latest"
    
   az vm create \
   --resource-group $RESOURCE_GROUP_NAME \
   --name $VM_NAME \
   --image $VM_IMAGE \
   --vnet-name $VNET_NAME \
   --subnet "MySubnet" \
   --public-ip-address "" \
   --nsg ""
    
   export IP_ADDRESS=$(az vm show --show-details --resource-group $RESOURCE_GROUP_NAME --name $VM_NAME --query privateIps --output tsv)
   ```

### Allow SAP BTP, Kyma Runtime to Peer with Your Remote Network

Tag the VPC network with the Kyma shoot name:

   ```shell
   export SHOOT_NAME=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}') 
   export VNET_ID=$(az network vnet show --name $VNET_NAME --resource-group $RESOURCE_GROUP_NAME --query id --output tsv)
   az tag update --resource-id $VNET_ID --operation Merge --tags $SHOOT_NAME
   ```

### Create VPC Peering

1. Create an AzureVpcPeering resource:

   ```shell
   kubectl apply -f - <<EOF
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AzureVpcPeering
   metadata:
     name: peering-to-my-vnet
   spec:
     remotePeeringName: peering-to-my-kyma
     remoteVnet: $VNET_ID
   EOF
   ```

2. Wait for the AzureVpcPeering to be in the `Ready` state.

   ```shell
   kubectl wait --for=condition=Ready azurevpcpeering/peering-to-my-vnet --timeout=300s
   ```

   Once the newly created AzureVpcPeering is provisioned, you should see the following message:

   ```console
   azurevpcpeering.cloud-resources.kyma-project.io/peering-to-my-vnet condition met
   ```

3. Create a namespace and export its value as an environment variable:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```

4. Create a workload that pings the VM in the remote network.

   ```shell
   kubectl apply -n $NAMESPACE -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: azurevpcpeering-demo
   spec:
     selector:
       matchLabels:
         app: azurevpcpeering-demo
     template:
       metadata:
         labels:
           app: azurevpcpeering-demo
       spec:
         containers:
         - name: my-container
           resources:
             limits:
               memory: 512Mi
               cpu: "1"
             requests:
               memory: 256Mi
               cpu: "0.2"
           image: ubuntu
           command:
             - "/bin/bash"
             - "-c"
             - "--"
           args:
             - "apt update; apt install iputils-ping -y; ping -c 20 $IP_ADDRESS"
   EOF
   ```

   This workload should print a sequence of 20 echo replies to stdout.

5. To print the logs of one of the workloads, run:

   ```shell
   kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=azurevpcpeering-demo -o=jsonpath='{.items[0].metadata.name}'`
   ```

   The command prints an output similar to the following:

   ```console
   ...
   PING 172.0.0.4 (172.0.0.4) 56(84) bytes of data.
   64 bytes from 172.0.0.4: icmp_seq=1 ttl=63 time=8.10 ms
   64 bytes from 172.0.0.4: icmp_seq=2 ttl=63 time=2.01 ms
   64 bytes from 172.0.0.4: icmp_seq=3 ttl=63 time=7.02 ms
   64 bytes from 172.0.0.4: icmp_seq=4 ttl=63 time=1.87 ms
   64 bytes from 172.0.0.4: icmp_seq=5 ttl=63 time=1.89 ms
   64 bytes from 172.0.0.4: icmp_seq=6 ttl=63 time=4.75 ms
   64 bytes from 172.0.0.4: icmp_seq=7 ttl=63 time=2.01 ms
   64 bytes from 172.0.0.4: icmp_seq=8 ttl=63 time=4.26 ms
   64 bytes from 172.0.0.4: icmp_seq=9 ttl=63 time=1.89 ms
   64 bytes from 172.0.0.4: icmp_seq=10 ttl=63 time=2.08 ms
   64 bytes from 172.0.0.4: icmp_seq=11 ttl=63 time=2.01 ms
   64 bytes from 172.0.0.4: icmp_seq=12 ttl=63 time=2.24 ms
   64 bytes from 172.0.0.4: icmp_seq=13 ttl=63 time=1.80 ms
   64 bytes from 172.0.0.4: icmp_seq=14 ttl=63 time=4.32 ms
   64 bytes from 172.0.0.4: icmp_seq=15 ttl=63 time=2.03 ms
   64 bytes from 172.0.0.4: icmp_seq=16 ttl=63 time=2.03 ms
   64 bytes from 172.0.0.4: icmp_seq=17 ttl=63 time=5.19 ms
   64 bytes from 172.0.0.4: icmp_seq=18 ttl=63 time=1.86 ms
   64 bytes from 172.0.0.4: icmp_seq=19 ttl=63 time=1.92 ms
   64 bytes from 172.0.0.4: icmp_seq=20 ttl=63 time=1.92 ms
    
   === 172.0.0.4 ping statistics ===
   20 packets transmitted, 20 received, 0% packet loss, time 19024ms
   rtt min/avg/max/mdev = 1.800/3.060/8.096/1.847 ms
   ...
   ```

## Next Steps

To clean up Kubernetes resources and your subscription resources, follow these steps:

1. Remove the created workloads:

   ```shell
   kubectl delete -n $NAMESPACE deployment azurevpcpeering-demo
   ```

2. Remove the created AzureVpcPeering resource:

    ```shell
    kubectl delete -n $NAMESPACE azurevpcpeering peering-to-my-vnet
    ```

3. Remove the created namespace:

    ```shell
    kubectl delete namespace $NAMESPACE
    ```

4. In your Microsoft Azure account, remove the created Azure resource group:

    ```shell
    az group delete --name $RESOURCE_GROUP_NAME --yes
    ```
