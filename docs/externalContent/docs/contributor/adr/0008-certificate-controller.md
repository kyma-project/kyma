# Certificate Management for APIRule Conversion Webhook with a New Controller

## Status

Proposed

## Context

We introduce a conversion webhook that converts APIRule versions from `v1beta1` to `v2alpha1` and vice versa. It requires a valid x509 certificate for the webhook server. We must provide implementation of certificate handling, which includes tasks such as creating, verifying, and renewing the certificate. Previously, for the conversion webhook, we used a CronJob to schedule a Job that would periodically handle certificate verification and renewal for the conversion webhook. However, this approach required us to maintain, build, and release an image for the webhook.

## Decision

1. We integrate a new Kubernetes controller into the module's operator to handle the creation, verification, and renewal of the required certificate.
2. We do not create an additional image specifically for handling certificate management using a CronJob. This approach has proven to have a few limitations, such as issues with updating the Job and inconveniences in regards to observability.
3. The new controller will reconcile predefined Secret named `api-gateway-webhook-certificate` in the `kyma-system` namespace. This Secret contains the data for the Certificate.
4. We add a Kubernetes `init container` to the operator Deployment. The container will handle the initial creation of the predefined Secret that holds the Certificate.
5. The created Secret `api-gateway-webhook-certificate` will have an **OwnerReference** set to the `api-gateway-manager` Deployment for cascading deletion.
6. We delegate function to the webhook server for obtaining the current Certificate. This will fully automate the Certificate renewal process.
7. We rotate the Certificate 14 days before expiration, and we create a Certificate with 90 days validity. SAP's recommendation for SSL server certificates is one year of validity.

## Consequences

The APIRule conversion works out of the box with an integrated conversion webhook started with the controller manager. We manage the certificate in our module's operator falling Kubernetes controller reconciling pattern.

The `api-gateway-webhook-certificate` Secret is not automatically recreated if deleted manually. To restore the secret, a restart of the manager is necessary.
