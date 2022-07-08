#!/usr/bin/env bash

TEST_ORY_IMAGE="eu.gcr.io/kyma-project/incubator/test-hydra-login-consent:d6e6d3bc"

function check_required_envs() {
    if [[ -z ${CLUSTER_NAME} ]]; then
            echo "You need to export CLUSTER_NAME as in CLUSTER_NAME.GARDENER_KYMA_PROW_PROJECT_NAME.shoot.canary.k8s-hana.ondemand.com" 
            exit 
    fi
    if [[ -z ${GARDENER_KYMA_PROW_PROJECT_NAME} ]]; then
            echo "You need to export GARDENER_KYMA_PROW_PROJECT_NAME as in CLUSTER_NAME.GARDENER_KYMA_PROW_PROJECT_NAME.shoot.canary.k8s-hana.ondemand.com" 
            exit 
    fi
}

function configure_ory_hydra() {
  echo "Prepare test environment variables"

  kubectl -n kyma-system set env deployment ory-hydra LOG_LEAK_SENSITIVE_VALUES="true"
  kubectl -n kyma-system set env deployment ory-hydra URLS_LOGIN="https://ory-hydra-login-consent.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com/login"
  kubectl -n kyma-system set env deployment ory-hydra URLS_CONSENT="https://ory-hydra-login-consent.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com/consent"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_ISSUER="https://oauth2.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com/"
  kubectl -n kyma-system set env deployment ory-hydra URLS_SELF_PUBLIC="https://oauth2.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com/"
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
  - ory-hydra-login-consent.${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.canary.k8s-hana.ondemand.com
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
}

check_required_envs
deploy_login_consent_app
configure_ory_hydra