#!/bin/bash
kubectl -n kyma-system set env deployment ory-hydra LOG_LEAK_SENSITIVE_VALUES="true"
kubectl -n kyma-system set env deployment ory-hydra URLS_LOGIN="https://ory-hydra-login-consent.kyma.example.com/login"
kubectl -n kyma-system set env deployment ory-hydra URLS_CONSENT="https://ory-hydra-login-consent.kyma.example.com/consent"
kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_ISSUER="https://oauth2.kyma.example.com/"
kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_PUBLIC="https://oauth2.kyma.example.com/"
kubectl -n kyma-system rollout restart deployment ory-hydra
kubectl -n kyma-system rollout status deployment ory-hydra
