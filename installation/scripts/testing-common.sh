ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${ROOT_PATH}/utils.sh

function context_arg() {
    if [ -n "$KUBE_CONTEXT" ]; then
        echo "--context $KUBE_CONTEXT"
    fi
}

function cmdGetPodsForSuite() {
    local suiteName=$1
    cmd="kubectl $(context_arg) get pods -l testing.kyma-project.io/suite-name=${suiteName} \
            --all-namespaces \
            --no-headers=true \
            -o=custom-columns=name:metadata.name,ns:metadata.namespace"
    echo $cmd
}

function printLogsFromFailedTests() {
    local suiteName=$1
    cmd=$(cmdGetPodsForSuite $suiteName)

    pod=""
    namespace=""
    idx=0

    for podOrNs in $($cmd)
    do
        n=$((idx%2))
         if [[ "$n" == 0 ]];then
            pod=${podOrNs}
            idx=$((${idx}+1))
            continue
        fi
        namespace=${podOrNs}
        idx=$((${idx}+1))

        log "Testing '${pod}' from namespace '${namespace}'" nc bold

        phase=$(kubectl $(context_arg)  get pod ${pod} -n ${namespace} -o jsonpath="{ .status.phase }")

        case ${phase} in
        "Failed")
            log "'${pod}' has Failed status" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        "Running")
            log "'${pod}' failed due to too long Running status" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        "Pending")
            log "'${pod}' failed due to too long Pending status" red
            printf "Fetching events from '${pod}':\n"
            kubectl $(context_arg)  describe po ${pod} -n ${namespace} | awk 'x==1 {print} /Events:/ {x=1}'
        ;;
        "Unknown")
            log "'${pod}' failed with Unknown status" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        "Succeeded")
            # do nothing
        ;;
        *)
            log "Unknown status of '${pod}' - ${phase}" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        esac
        log "End of testing '${pod}'\n" nc bold

    done
}

function getContainerFromPod() {
    local namespace="$1"
    local pod="$2"
    local containers2ignore="istio-init istio-proxy manager"
    containersInPod=$(kubectl get pods ${pod} -o jsonpath='{.spec.containers[*].name}' -n ${namespace})
    for container in $containersInPod; do
        if [[ ! ${containers2ignore[*]} =~ "${container}" ]]; then
            echo "${container}"
        fi
    done
}

function printLogsFromPod() {
    local namespace=$1 pod=$2
    local tailLimit=2000 bytesLimit=500000
    log "Fetching logs from '${pod}' with options tailLimit=${tailLimit} and bytesLimit=${bytesLimit}" nc bold
    testPod=$(getContainerFromPod ${namespace} ${pod})
    result=$(kubectl $(context_arg)  logs --tail=${tailLimit} --limit-bytes=${bytesLimit} -n ${namespace} -c ${testPod} ${pod})
    if [ "${#result}" -eq 0 ]; then
        log "FAILED" red
        return 1
    fi
    echo "${result}"
}

function checkTestPodTerminated() {
    local suiteName=$1
    runningPods=false

    pod=""
    namespace=""
    idx=0

    cmd=$(cmdGetPodsForSuite $suiteName)
    for podOrNs in $($cmd)
    do
       n=$((idx%2))
       if [[ "$n" == 0 ]];then
         pod=${podOrNs}
         idx=$((${idx}+1))
         continue
       fi
        namespace=${podOrNs}
        idx=$((${idx}+1))

        phase=$(kubectl $(context_arg)  get pod "$pod" -n ${namespace} -o jsonpath="{ .status.phase }")
        # A Pod's phase  Failed or Succeeded means pod has terminated.
        # see: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
        if [ "${phase}" !=  "Succeeded" ] && [ "${phase}" != "Failed" ]
        then
          log "Test pod '${pod}' has not terminated, pod phase: ${phase}" red
          runningPods=true
        fi
    done

    if [ ${runningPods} = true ];
    then
        return 1
    fi
}

function waitForTestPodsTermination() {
    local retry=0
    local suiteName=$1

    log "All test pods should be terminated. Checking..." nc bold
    while [ ${retry} -lt 3 ]; do
        checkTestPodTerminated ${suiteName}
        checkTestPodTerminatedErr=$?
        if [ ${checkTestPodTerminatedErr} -ne 0 ]; then
            echo "Waiting for test pods to terminate..."
            sleep 1
        else
            log "OK" green bold
            return 0
        fi
        retry=$[retry + 1]
    done
    log "FAILED" red
    return 1
}

function waitForTerminationAndPrintLogs() {
    local suiteName=$1

    waitForTestPodsTermination ${suiteName}
    checkTestPodTerminatedErr=$?

    printLogsFromFailedTests ${suiteName}
    if [ ${checkTestPodTerminatedErr} -ne 0 ]
    then
        return 1
    fi
}

function printImagesWithLatestTag() {

    local images=$(kubectl $(context_arg)  get pods --all-namespaces -o jsonpath="{..image}" |\
    tr -s '[[:space:]]' '\n' |\
    grep ":latest")

    log "Images with tag latest are not allowed. Checking..." nc bold
    if [ ${#images} -ne 0 ]; then
        log "${images}" red
        log "FAILED" red
        return 1
    fi
    log "OK" green bold
    return 0
}

TESTING_ADDONS_CFG_NAME="testing-addons"

function injectTestingBundles() {
    cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  labels:
    controller-tools.k8s.io: "1.0"
  name: clusteraddonsconfigurations.addons.kyma-project.io
spec:
  group: addons.kyma-project.io
  names:
    categories:
    - all
    - addons
    kind: ClusterAddonsConfiguration
    plural: clusteraddonsconfigurations
  scope: Cluster
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            repositories:
              items:
                properties:
                  url:
                    type: string
                required:
                - url
                type: object
              type: array
            reprocessRequest:
              format: int64
              type: integer
          required:
          - repositories
          type: object
        status:
          properties:
            lastProcessedTime:
              format: date-time
              type: string
            observedGeneration:
              format: int64
              type: integer
            phase:
              type: string
            repositories:
              items:
                properties:
                  addons:
                    items:
                      properties:
                        message:
                          type: string
                        name:
                          type: string
                        reason:
                          type: string
                        status:
                          enum:
                          - Ready
                          - Failed
                          type: string
                        version:
                          type: string
                      required:
                      - name
                      - version
                      type: object
                    type: array
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  url:
                    type: string
                required:
                - url
                - addons
                type: object
              type: array
          required:
          - phase
          type: object
  version: v1alpha1
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
EOF

    cat <<EOF | kubectl apply -f -
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  labels:
    addons.kyma-project.io/managed: "true"
  name: ${TESTING_ADDONS_CFG_NAME}
spec:
  repositories:
  - url: "https://github.com/kyma-project/bundles/releases/download/0.6.0/index-testing.yaml"
EOF
    log "Testing addons injected" green
}

function removeTestingBundles() {
    kubectl delete clusteraddonsconfiguration ${TESTING_ADDONS_CFG_NAME}
    log "Testing addons removed" green
}