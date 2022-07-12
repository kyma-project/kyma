#!/usr/bin/env bash

TEST_ORY_IMAGE="eu.gcr.io/kyma-project/incubator/test-hydra-login-consent:d6e6d3bc"

function check_required_envs() {
    if [[ -z ${KYMA_DOMAIN} ]]; then
            echo "You need to export KYMA_DOMAIN"
            exit 
    fi
}

function configure_ory_hydra() {
  echo "Prepare test environment variables"

  kubectl -n kyma-system set env deployment ory-hydra LOG_LEAK_SENSITIVE_VALUES="true"
  kubectl -n kyma-system set env deployment ory-hydra URLS_LOGIN="https://ory-hydra-login-consent.${KYMA_DOMAIN}/login"
  kubectl -n kyma-system set env deployment ory-hydra URLS_CONSENT="https://ory-hydra-login-consent.${KYMA_DOMAIN}/consent"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_ISSUER="https://oauth2.${KYMA_DOMAIN}/"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_PUBLIC="https://oauth2.${KYMA_DOMAIN}/"
  kubectl -n kyma-system rollout restart deployment ory-hydra
  kubectl -n kyma-system rollout status deployment ory-hydra
}

function deploy_login_consent_app() {
  echo "Deploying Ory login consent app for tests"

  kubectl -n istio-system rollout status deployment istiod
  kubectl -n istio-system rollout status deployment istio-ingressgateway

  kubectl apply -f "$PWD/ory-hydra-login-consent.yaml"
  echo "App deployed"
  rm "$PWD/ory-hydra-login-consent.yaml"
}

check_required_envs
deploy_login_consent_app
configure_ory_hydra