# XIP patch

## Overview

This component configures Kyma to use a wildcard DNS provided by `xip.io`.

## Prerequisites

This component must be installed after `cluster-essentials` the component and before `istio-kyma-patch`.

## Usage

You can't configure this component through environment variables.

This component performs the following actions:
 1. Reads the external address of the`istio-ingressgateway` service. 
 2. Sets the **global.domainName** to `{IP_ADDRESS}.xip.io`.
 3. Creates a self-signed certificate for this domain.
 4. Sets the **global.tlsCrt** and **global.tlsKey** to use the created certificate.
