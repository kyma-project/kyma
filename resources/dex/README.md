# Dex

## Overview

Dex is an identity service that delegates user authentication to external identity providers using [connectors](https://github.com/coreos/dex#connectors).
For more details about Dex, see the [Dex GitHub](https://github.com/coreos/dex) project.

## Details

Currently, Dex uses a static user database and authenticates static users by itself, instead of using a fully-integrated authentication solution. Dex also comes with a static list of clients allowed to initiate the OAuth2 flow.

For the list of static Dex users and clients, as well as the information about the connectors that delegate authentication to external identity providers, see the [dex-config-map.yaml](templates/dex-config-map.yaml) file.

Dex is exposed using the [Istio VirtualService](https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/) feature. Access Dex at `https://dex.kyma.local`.

## Configuration

This chart allows to provide configuration for Dex connectors and clients using Helm overrides mechanism.


### Connectors

Connectors can be configured using override named `connectors`.
The value of the override must be a single string containing Dex connectors configuration in YAML format. See [Dex connectors documentation](https://github.com/dexidp/dex/tree/master/Documentation/connectors) for details.
Note you can use Go Template expressions inside `connectors` override. These expressions will be resolved by Helm using the same set of overrides as configured for the entire chart.

Example:
```
  connectors: |-
    - type: saml
      id: sci
      name: SAP CI
      config:
        # Issuer for SAML Request
        entityIssuer: dex.{{ .Values.global.domainName }}
        ssoURL: https://{{ .Values.idp.tenant | default "someDefault" }}.{{ .Values.idp.domain | default "mytenant.mydomain.com" }}/saml2/idp/sso?sp=dex.{{ .Values.global.domainName }}
        ca: {{ .Values.idp.caPath}}/ca.pem
        redirectURI: https://dex.{{ .Values.global.domainName }}/callback
        usernameAttr: mail
        emailAttr: mail
        groupsAttr: groups
```

### Clients
Configuration for static clients is split in two parts: `staticClientsBase` and `staticClientsExtra`
The `staticClientsBase` clients are basic clients required by Kyma and should not be modified.
Users can provide additional clients using Helm override named `oidc.staticClientsExtra`.
The value of this override is a string in YAML format with a list of clients.
Note you can use Go Template expressions in the override value. These expressions will be resolved by Helm using the same set of overrides as configured for the entire chart.

Example:
```
  oidc.staticClientsExtra: |-
    - id: console2
      name: Console2
      redirectURIs:
      - 'http://console-dev.{{ .Values.global.ingress.domainName }}:4200'
      - 'https://console.{{ .Values.global.ingress.domainName }}'
      secret: a1b2c3d4xyz
```
