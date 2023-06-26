I would like to unify naming at least in serverless.

So let's start
Currently we have components
- serverless, which is responsible for running functions
- serverless-manager which is resposnbile for isntalation and configuration of serverless
- keda which is install by our operator 

Very often we use name controller to name:
- serverless reconcile loop
- serverless pod with reconcile loop

I would like to propose naming convention for serverless

* Serverless - name of the component which has defintion of CR and runs, configure functions
* Serverless-operator - name of the component which install and configure Serverless
* Kyma-Keda-operator

Serverless consist of two services:

- reconcile loop
- defaulting/validation/conversion webhooks

First component is responsible for creating, configuring k8s resources to finally run function on a cluster
So, the main responsibility is reconcling functions cr

Webhook parts is resposnbile for defaulting/validation for function CR, mutating external registry secret but also reconcile certificates.
The main responsibility is doing webhook things. 

The kubebuilder describe its architecture like in following picture.

TODO: daj tutaj zdjęcie z kubebuildera

We can see that `Process` (generaly pod) have manager which is the brain of the ?component?
The manager can have 2 sub-components:
- Controller, which focuses on reconciliation of given k8s resoruce. Has predicate and reconciler.
- Webhook, which works with AdmissionRequests

My proposal about naming:

for pod with reconcile loop: 
- ${component_name}-controller
- ${component_name}-{crd_name}-controller in case of having seperate reconile loop

for pod with webhook: ${component_name}-webhook
for pod with both of those components: ${component_name}

for reconcile loop: 
- ${component_name}-${crd_name}-reconcile
- ${crd_name}-reconcile

for validation webhook: 
- ${component_name}-${crd_name}-validaton-webhook
- ${component_name}-validaton-webhook
- ${crd_name}-validaton-webhook
