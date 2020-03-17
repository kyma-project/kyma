#!/bin/bash -e

# 1 - Get current secrets (from mounted secret)
# 2 - Get current overrides (kubectl get secret -n kyma-installer -l "installer=overrides,component=ory")
LOCAL_SECRETS=$(ls /etc/secrets)
OVERRIDE_CM=$(kubectl get cm -n kyma-installer -l "installer=overrides,component=ory")
OVERRIDE_SECRET=$(kubectl get secret -n kyma-installer -l "installer=overrides,component=ory")

# kubectl get cm -n kyma-installer -l "installer=overrides,component=ory" -o jsonpath='{.items[].data}'

echo $LOCAL_SECRETS
echo $OVERRIDE_SECRET
echo $OVERRIDE_CM