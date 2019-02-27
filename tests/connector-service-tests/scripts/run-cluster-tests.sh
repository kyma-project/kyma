#set -e

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"


if [[ -z "${TEST_IMAGE}" ]]; then
  echo "TEST_IMAGE env is not set. It should be set to full path of and image including tag, ex: mydockerhub/connector-service-tests:0.0.1"
  exit 1
fi

if [[ -z "${DOMAIN}" ]]; then
  echo "DOMAIN_NAME env is not set. It should be set to cluster domain name, ex: nightly.cluster.kyma.cx"
  exit 1
fi

echo "Current cluster context: $(kubectl config current-context)"

echo "Image: $TEST_IMAGE"
echo "Domain: $DOMAIN"

echo ""
echo "------------------------"
echo "Removing test pod"
echo "------------------------"


kubectl -n kyma-integration delete po connector-service-tests --now

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t $TEST_IMAGE

echo ""
echo "------------------------"
echo "Push tests image"
echo "------------------------"

docker push $TEST_IMAGE

echo ""
echo "------------------------"
echo "Creating test pod"
echo "------------------------"

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: connector-service-tests
  annotations:
    sidecar.istio.io/inject: "false"
spec:
  containers:
  - name: connector-service-tests
    image: $TEST_IMAGE
    imagePullPolicy: Always
    env:
    - name: INTERNAL_API_URL
      value: http://connector-service-internal-api:8080
    - name: EXTERNAL_API_URL
      value: http://connector-service-external-api:8081
    - name: GATEWAY_URL
      value: https://gateway.$DOMAIN
    - name: SKIP_SSL_VERIFY
      value: "true"
    - name: CENTRAL
      value: "true"
  restartPolicy: Never
EOF

echo ""
echo "------------------------"
echo "Waiting 10 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 10

kubectl -n kyma-integration logs connector-service-tests -f -c connector-service-tests
