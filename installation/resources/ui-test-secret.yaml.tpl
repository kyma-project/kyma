apiVersion: v1
kind: Secret
metadata:
  name: ui-test
  namespace: kyma-installer
type: Opaque
data:
  user: "__UI_TEST_USER__"
  password: "__UI_TEST_PASSWORD__"
