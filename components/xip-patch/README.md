# Istio Kyma patch

## Overview

This component configures Kyma to use a wildcard DNS provided by `xip.io`.

## Prerequisites

This component must be installed after `cluster-essentials` the component and before `istio-kyma-patch`.

## Usage

You can't configure this component through environment variables.

This component performs the following actions:
 1. Reads `istio-ingressgateway` Service's external address.
 2. Sets `global.domainName` to `{ip address}.xip.io`.
 3. Creates self-signed certificate for this domain.
 4. Sets `global.tlsCrt` and `global.tlsKey` to contain created certificate.
