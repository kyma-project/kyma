---
title: Certificate management
type: Details
---

Kyma uses a bundled certificate management solution, [cert-manager](https://cert-manager.io/).
You can use its functionality to generate and manage TLS certificates.

Out-of the box Kyma supports four predefined modes for handling certificates:
- Gardener
- xip.io
- Legacy
- User-Provided

The xip mode has to be enabled manually, the other modes are auto-detected.
You also can disable built-in cert-manager by setting a global [override](link to overrides docs) `global.certificates.manager.enabled: "false"`.

## Modes

### Legacy mode

This mode is only supported for local installations.
It is detected when installation overrides containing certificate data named `global.tlsCrt` and `global.tlsKey` are defined.
In this mode cert-manager is not used to manage certificates.
You can manually enable this mode by defining an installation override: `global.certificates.type: "legacy"`.

### Gardener mode

Gardener mode is detected based on the existence of Gardener-specific api-version: `cert.gardener.cloud`.
In this mode the cert-manager is not used to manage Kyma certificates. The Gardener issues certificates for your services based on annotations.
You can manually enable this mode by defining an installation override: `global.certificates.type: "gardener"`.

### Xip mode

In xip mode, cert-manager is used to create self-signed certificates for xip.io managed domains.
You have to manually enable this mode by defining an installation override: `global.certificates.type: "xip"`.
Because it's a self-signed certificate, you have to manually mark it as a trusted certifcate, either in your browser of host OS (here will be a link to how-to-do-it).

### User-provided mode

User-provided mode is a default one when no other mode is detected.
You can manually enable this mode by defining an installation override: `global.certificates.type: "user-provided"`.
This mode allows you to directly use cert-manager features. You have to manually create cert-manager custom resources such as Issuer and Certificate.

> **NOTE:** In this setup, users provide their own issuers, which can be globally trusted (like LetsEncrypt), or SelfSigned. 
Depending on the chosen flow and issuer, You may need credentials that allow cert-manager to set up DNS records for issuing the certificate. Refer to [cert-manager documentation](https://cert-manager.io/) for details.
