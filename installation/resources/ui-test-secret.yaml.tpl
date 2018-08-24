apiVersion: v1
kind: Secret
metadata:
  name: ui-test
  namespace: kyma-installer
type: Opaque
data:
  user: "__UI_TEST_USER__"
  password: "__UI_TEST_PASSWORD__"
---
apiVersion: v1
kind: Secret
metadata:
  name: ui-test-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  test.auth.username: "__UI_TEST_USER__"
  test.auth.password: "__UI_TEST_PASSWORD__"