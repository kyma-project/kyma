# Dex

## Overview

Dex is an identity service that delegates user authentication to external identity providers using [connectors](https://github.com/coreos/dex#connectors).
For more details about Dex, see the [Dex GitHub](https://github.com/coreos/dex) project.

## Details

Currently, Dex uses a static user database and authenticates static users by itself, instead of using a fully-integrated authentication solution. Dex also comes with a static list of clients allowed to initiate the OAuth2 flow.

For the list of static Dex users and clients, as well as the information about the connectors that delegate authentication to external identity providers, see the [dex-config-map.yaml](templates/dex-config-map.yaml) file.

Dex is exposed using the [Istio VirtualService](https://istio.io/docs/reference/config/networking/virtual-service/) feature. Access Dex at `https://dex.{CLUSTER_DOMAIN}`.

## Configuration

This chart allows you to provide configuration for Dex connectors and clients using the Helm overrides mechanism.

>**TIP:** You can use Go Template expressions in the override value. These expressions are resolved by Helm using the same set of overrides as configured for the entire chart.

### Connectors

Configure connectors through the `connectors` override.
Provide the Dex connectors list as a single string in the `yaml` format. See [these](https://github.com/dexidp/dex/tree/master/Documentation/connectors) documents for syntax details.

This is an example of a connector configuration string:
```yaml
  connectors: |-
    - type: saml
      id: iaa
      name: IAA
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

Configure Dex clients through the `oidc.staticClientsExtra` override. Pass the list of clients as a single string in the `yaml` format.

>**CAUTION:** The `oidc.staticClientsBase` override defines the basic clients required by Kyma. Do not edit this override.

This is an example of a client configuration string:
```yaml
  oidc.staticClientsExtra: |-
    - id: console2
      name: Console2
      redirectURIs:
      - 'http://console-dev.{{ .Values.global.ingress.domainName }}:4200'
      - 'https://console.{{ .Values.global.ingress.domainName }}'
      secret: a1b2c3d4xyz
```
### Custom volumes

Configure additional volumes and mounts required for certificates using the `volumeMountsExtra` and `volumesExtra` overrides. Pass the list of volumes and mounts as a single string in the `yaml` format.

This is an example of an extra volume and mount pair:
```yaml
volumeMountsExtra: |-
  - name: config
    mountPath: /foo
volumesExtra: |-
  - name: extra-config
    emptyDir: {}
```
