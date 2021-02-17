# XIP patch

## Overview

This component configures Kyma to use a wildcard DNS provided by `xip.io`.

## Prerequisites

This component must be installed after the `cluster-essentials` component.

## Usage

The patch accepts the **EXTERNAL_PUBLIC_IP** environment variable which must contain a manually reserved, external IP address. If this 
variable is not set, the script tries to get the external IP address from the `istio-ingressgateway` service.

This component performs the following actions:
 1. Reads the external IP address. 
 2. Sets the **global.domainName** to `{IP_ADDRESS}.xip.io`.
 3. Creates a self-signed certificate for this domain.
 4. Sets the **global.tlsCrt** and **global.tlsKey** to use the created certificate.
