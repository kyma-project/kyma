---
title: Identity Provider Presets
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

| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.issuer** | **YES** | Specifies the issuer of the JWT tokens used to access the services. |
| **spec.jwksUri** | **YES** | Specifies the URL of the OpenID Provider’s public key set to validate the signature of the JWT token. |

## Usage in the UI

**issuer** and **jwksUri** are some of the API CR specification fields. In most cases, these values are reused many times. IDPPresets usage is a solution to reuse them in a convenient way. It allows you to choose a proper preset from the dropdown menu instead of entering them manually every time you expose a secured API. Apart from consuming the IDPPresets, you can also manage them in the Console UI. To create and delete IDPPresets, go to the **Administration** tab and then to **IDP Presets**.

## Related resources and components

These components use this CR:

| Name   |   Description |
|:----------:|:------|
| IDP Preset |  Generates Go client which allows components and tests to create, delete, or get IDP Preset resources. |
| UI API Layer |  Enables the IDP Preset management with GraphQL API. |
