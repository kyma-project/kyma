apiVersion: v1
kind: ConfigMap
metadata:
  name: dockerfile-${RUNTIME_NAME}
  namespace: kyma-system
  labels:
    serverless.kyma-project.io/config: runtime
    serverless.kyma-project.io/runtime: ${RUNTIME_NAME}
data:
  Dockerfile: |-
${DOCKERFILE}
