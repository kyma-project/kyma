# Webhook Certificate Initialization

## Overview

This application can be used to initialize webhooks with a certificate and key. Best used as an Init Container in the Operator Pods where the certificate is shared by the `emptyDir` Volume.

## Development

### Available Commands

For development, use the following commands:

- Run all tests and validation:

```bash
make
```

- Run locally:

```bash
make run-local
```
