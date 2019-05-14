---
title: IDPPreset
type: Custom Resource
---

The `idppresets.authentication.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format that represents presets of the Identity Provider configuration used to secure API through the Console UI. Presets are a convenient way to configure the **authentication** section in the API custom resource.

To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd idppresets.authentication.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR used to create an IDPPreset:

```yaml
apiVersion: authentication.kyma-project.io/v1alpha1
kind: IDPPreset
metadata:
    name: "sample-idppreset"
spec:
    issuer: https://example.com
    jwksUri: https://example.com/keys
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |  Mandatory  |  Description |
|----------|:-------------:|------|
| **metadata.name** | **YES** | Specifies the name of the CR. |
| **spec.issuer** | **YES** | Specifies the issuer of the JWT tokens used to access the services. |
| **spec.jwksUri** | **YES** | Specifies the URL of the OpenID Provider’s public key set to validate the signature of the JWT token. |

## Usage in the UI

The **issuer** and **jwksUri** fields originate from the [Api CR](/components/api-gateway/#custom-resource-custom-resource) specification. In most cases, these values are reused many times. Use the IDPPreset CR to store these details in a single object and reuse them in a convenient way. In the UI, the IDPPreset CR allows you to choose a preset with details of a specific identity provider from the drop-down menu instead of entering them manually every time you expose a secured API. Apart from consuming IDPPresets, you can also manage them in the Console UI. To create and delete IDPPresets, select **IDP Presets** from the **Integration** section.

## Related resources and components

These components use this CR:

| Name   |   Description |
|----------|------|
| IDP Preset |  Generates a Go client which allows components and tests to create, delete, or get IDP Preset resources. |
| Console Backend Service |  Enables the IDP Preset management with GraphQL API. |
