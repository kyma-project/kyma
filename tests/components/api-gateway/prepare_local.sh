#!/usr/bin/env bash

TEST_ORY_IMAGE="eu.gcr.io/kyma-project/incubator/test-hydra-login-consent:d6e6d3bc"

function check_required_envs() {
    if [[ -z ${KYMA_DOMAIN} ]]; then
            echo "KYMA_DOMAIN not exported, fallback to default k3d local.kyma.dev"
            export KYMA_DOMAIN=local.kyma.dev
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

cat << EOF > "$PWD/ory-hydra-login-consent.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ory-hydra-login-consent
  namespace: kyma-system
spec:
  selector:
    matchLabels:
      app: ory-hydra-login-consent
      version: v1
  template:
    metadata:
      labels:
        app: ory-hydra-login-consent
        version: v1
    spec:
      containers:
        - name: login-consent
          image: ${TEST_ORY_IMAGE}
          env:
            - name: HYDRA_ADMIN_URL
              value: http://ory-hydra-admin.kyma-system.svc.cluster.local:4445
            - name: BASE_URL
              value: ""
            - name: PORT
              value: "3000"
          ports:
          - containerPort: 3000
---
kind: Service
apiVersion: v1
metadata:
  name: ory-hydra-login-consent
  namespace: kyma-system
spec:
  selector:
    app: ory-hydra-login-consent
    version: v1
  ports:
    - name: http-login-consent
      protocol: TCP
      port: 80
      targetPort: 3000
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: ory-hydra-login-consent
  namespace: kyma-system
  labels:
    app: ory-hydra-login-consent
spec:
  gateways:
  - kyma-system/kyma-gateway
  hosts:
  - ory-hydra-login-consent.${KYMA_DOMAIN}
  http:
  - match:
    - uri:
        exact: /login
    - uri:
        exact: /consent
    route:
    - destination:
        host: ory-hydra-login-consent.kyma-system.svc.cluster.local
        port:
          number: 80
EOF
  kubectl apply -f "$PWD/ory-hydra-login-consent.yaml"
  echo "App deployed"
  rm "$PWD/ory-hydra-login-consent.yaml"
}

function update_mtls_gateway_cacert_secret() {
client_root_ca_crt_file='assets/kyma-root-ca.crt'
#client_crt_file='assets/testmtls.kyma-example.com.crt'
#client_key_file='assets/testmtls.kyma-example.com.key'
clientRootCAEncoded=$(cat $client_root_ca_crt_file| base64)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: kyma-mtls-gateway-certs-cacert
  namespace: istio-system
type: Opaque
data:
  cacert: ${clientRootCAEncoded}
EOF
}

check_required_envs
deploy_login_consent_app
configure_ory_hydra
update_mtls_gateway_cacert_secret