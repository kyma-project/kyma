---
title: Identity Provider Presets
type: Custom Resource
---

The `idppresets.authentication.kyma-project.io` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format that represents presets of Identity Provider configuration used in securing API through console UI. Presets are a convenient way to configure authentication section in API Custom Resource.

To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd idppresets.authentication.kyma-project.io -o yaml
```

## Sample Custom Resource

This is a sample CR used to create a IDP Preset:

```yaml
apiVersion: authentication.kyma-project.io/v1alpha1
kind: IDPPreset
metadata:
    name: "sample-idppreset"
spec:
    issuer: https://example.com
    jwksUri: https://example.com/keys
    name: "sample-idppreset"
```

## Properties of Custom Resource

This table analyses the elements of the CR and the information it contains:

| Field   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.issuer** | **YES** | Specifies the issuer of the JWT tokens used to access the services |
| **spec.jwksUri** | **YES** | Specifies URL of the OpenID Provider’s public key set to validate signature of the JWT token. |
| **spec.name** | **YES** | Specifies the name of the preset. |

## Usage in UI

There are two functionalities related to IDPPresets in the Console UI: management of presets and utilising them. You can create and delete IDP Presets by going to `Administration` tab and then to `IDP Presets`. Ultilising was slightly described in the introduction. Issuer and jwksURI are some of the API CR specification fields. However, in most cases that values are reused many times. IDPPresets usage is a solution to reuse them in a convinient way, enabling user to choose proper preset from the dropdown menu instead of entering them manually every time the user exposes a secured API.