# Overview
A flexible and easy way to connect and extend enterprise applications in a cloud-native world

For more information on Kyma, see [kyma-project](https://kyma-project.io/).
## About Google Click to Deploy
Popular open source software stacks on Kubernetes packaged by Google and made available in Google Cloud Marketplace.
### Solution Information
This solution will deploy a Kyma instance to your kubernetes cluster. Kyma is installed through a Kyma installer container and requires no previous step.
# Installation
When you click on a GKE marketplace app, there are two options to install. Either with UI or with Command Line.

## Quick Install with Google Cloud Marketplace
Get up and running with a few clicks! Install this Kyma app to a Google Kubernetes Engine cluster using Google Cloud Marketplace. Follow the on-screen instructions.
## Command Line Instructions
### Prerequisites
#### Set up comand-line tools
You'll need the following tools in your development environment:
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)
- [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

#### Create a new cluster from command-line:
```sh
export GCP_PROJECT=<your-project-id>
export GCP_CLUSTER_NAME=<your-cluster-name>
export GCP_ZONE=europe-west3-a

#Create a cluster
gcloud container --project "$GCP_PROJECT" clusters \
        create "$GCP_CLUSTER_NAME" --zone "$GCP_ZONE" \
        --cluster-version "1.12" --machine-type "n1-standard-4" \
        --addons HorizontalPodAutoscaling,HttpLoadBalancing

```
#### Configure kubectl to connect to the cluster
```
gcloud container clusters get-credentials "$GCP_CLUSTER_NAME" --zone "$GCP_ZONE"
```
#### Clone this repo
```
git clone --recursive https://github.com/GoogleCloudPlatform/click-to-deploy.git
```
#### Install the Application resource definition
An Application resource is a collection of individual Kubernetes components, such as Services, Deployments, and so on, that you can manage as a group.

To set up your cluster to understand Application resources, run the following command:
```sh
kubectl apply -f "https://raw.githubusercontent.com/GoogleCloudPlatform/marketplace-k8s-app-tools/master/crd/app-crd.yaml"
```
You need to run this command once for each cluster.

The Application resource is defined by the Kubernetes SIG-apps community. The source code can be found on github.com/kubernetes-sigs/application.

#### Install the Kyma Application
Kyma requires cluster admin service account to operate correctly. To create one, run the following commands.
```sh
export SERVICE_ACCOUNT=kyma-serviceaccount
export NAMESPACE=default
export APPLICATION_NAME=my-uber-kyma-app
export KYMA_DEPLOYER_IMAGE=hudymi/kyma-deployer:1.1.1

kubectl create sa "$SERVICE_ACCOUNT" --namespace "$NAMESPACE"
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --serviceaccount="$NAMESPACE:$SERVICE_ACCOUNT"
Navigate to `Kyma` directory:
```sh
cd click-to-deploy/k8s/kyma
```

#### Expand the manifest template
```sh
awk 'FNR==1 {print "---"}{print}' manifest/* \
  | envsubst '$APPLICATION_NAME $NAMESPACE $SERVICE_ACCOUNT $KYMA_DEPLOYER_IMAGE' \
  > "${APPLICATION_NAME}_manifest.yaml"
```

#### Apply the manifest to your Kubernetes cluster
Use `kubectl` to apply the manifest to your Kubernetes cluster. This installation will create:
- An Application resource, which collects all the deployment resources into one logical entity.
- Kyma deployer

```sh
kubectl apply -f "${APPLICATION_NAME}_manifest.yaml" --namespace "${NAMESPACE}"
```
#### View the app in the Google Cloud Console
To get the console URL for your app, run the following command:
```sh
echo "https://console.cloud.google.com/kubernetes/application/${GCP_ZONE}/${GCP_CLUSTER_NAME}/${NAMESPACE}/${APPLICATION_NAME}?project=${GCP_PROJECT}"
```
To view the app, open the URL in your browser. To follow kyma installation follow the `workloads` panel in GKE UI.

Deployment flow looks like this:

![Deployment Flow](./resources/deployment.png)

## Using the app
### Accesing Kyma Console
Kyma uses self signed certificates for GKE deployment. To access kyma, you have to add those certificates to your keychain. Kyma certificate can be seen in Application panel of GKE. To add those certificate run the command for Windows:
```sh
echo '<cert>' | base64 -d  > kyma.cer
certutil -user -addstore Root kyma.cer
```
> If you don't have base64 command installed, use [online base64 decoder](https://www.base64decode.org/)
For OSX:
```sh
tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
&& kubectl get configmap cluster-certificate-overrides -n kyma-installer -o jsonpath='{.data.global\.tlsCrt}' | base64 --decode > $tmpfile \
&& sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
&& rm $tmpfile
```
You can see the address and access information of your console in your Application panel.

# Scaling
Scaling is done by the cluster's Horizontal Pod Autoscaler. Scaling is independent from core applications.

# Backup and Restore
For information about backing up your Kyma instance, see [kyma website](https://kyma-project.io/docs/).

# Updating the app
For information on updating Kyma, see [Updating Kyma](https://kyma-project.io/docs/root/kyma/#installation-update-kyma).

# Uninstall the Application
### Delete Resources
```sh
kubectl label installation/kyma-installation action=uninstall --overwrite=true
```
### Kyma application resource:
```sh
kubectl delete -f "${APP_INSTANCE_NAME}_manifest.yaml" --namespace "${NAMESPACE}"
```
### Delete GKE cluster
```sh
gcloud container clusters delete "$GKE_CLUSTER_NAME" 
```