# Cloud Event Gateway Proxy

## Overview


## Prerequisites


## Usage

Build

```bash
go mod vendor
```

Test

```bash
make test-local
```

Deploy

```bash
ko apply -f config/
```

## Environment Variables

| Environment Variable  | Description                                                                     |
| --------------------- | ------------------------------------------------------------------------------- |
| INGRESS_PORT          | The ingress port for the CloudEvents Gateway Proxy.                             |
| CLIENT_ID             | The Client ID used to acquire Access Tokens from the Authentication server.     |
| CLIENT_SECRET         | The Client Secret used to acquire Access Tokens from the Authentication server. |
| TOKEN_ENDPOINT        | The Authentication Server Endpoint to provide Access Tokens.                    |
| EMS_CE_URL            | The Messaging Server Endpoint that accepts publishing CloudEvents to it.        |

