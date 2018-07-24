#!/bin/bash

set -o errexit

FINAL_IMAGE="kyma-on-minikube"
ACR_BASE_URL="azurecr.io"
ACR_URL="${ACR_NAME}.${ACR_BASE_URL}"
POD_NAME="kyma-minikube-$IMAGE_TAG"
FULL_IMAGE_NAME="${ACR_URL}/${FINAL_IMAGE}:${IMAGE_TAG}"

function pushDockerImageToAcr {
  docker tag ${FINAL_IMAGE}:latest $FULL_IMAGE_NAME
  docker login ${ACR_NAME}.${ACR_BASE_URL} -u ${ARM_CLIENT_ID} -p ${ARM_CLIENT_SECRET}
  docker push $FULL_IMAGE_NAME
}

function setKubeconfig {
  echo $KUBECONFIG_JSON | base64 -d > kubeconfig
  export KUBECONFIG=kubeconfig
}

function deletePod {
  setKubeconfig
  
  kubectl delete pod $POD_NAME
}

function createPod {
  setKubeconfig

  cat <<EOF > pod.yaml
apiVersion: v1
kind: Pod
metadata:
 name: $POD_NAME
 labels:
   app: kyma
spec:
  restartPolicy: Never
  containers:
    - image: $FULL_IMAGE_NAME
      name: minikube
      resources:
        requests:
          memory: "6Gi"
      env:
      - name: RUN_TESTS
        value: "true"
      - name: IGNORE_TEST_FAIL
        value: "false"
      securityContext:
        privileged: true
      volumeMounts:
      - mountPath: /var/lib/docker
        name: docker-volume
      - mountPath: /var/lib/localkube/etcd
        name: etcd-volume
  volumes:
  - name: docker-volume
    emptyDir: {}
  - name: etcd-volume
    emptyDir: {}
EOF

  kubectl create -f pod.yaml

  for i in {1..20};
  do
    podStatus=$(kubectl get pod $POD_NAME -o jsonpath='{.status.phase}')
    if [[ "$podStatus" == "Running"  ]];
    then
      break
    fi
    containerReason=$(kubectl get pod $POD_NAME -o jsonpath='{.status.containerStatuses[0].state.waiting.message}')
    echo "Pod is in status: $podStatus - $containerReason. Waiting..."
    sleep 5
  done
  
  kubectl logs $POD_NAME -f

  podStatus=$(kubectl get pod $POD_NAME -o jsonpath='{.status.phase}')

  if [[ "$podStatus" == "Succeeded" ]];
  then
    echo "BUILD SUCCESS"
  else
    echo "BUILD FAILED"
    exit 1
  fi
}

# Run function passed as first argument
$1
