#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

CONCURRENCY=1

POSITIONAL=()

function validateConcurrency() {
  if [[ -z "$1" ]]; then
    echo "Error: --concurrency requres a value"
    exit 1
  fi

  if ! [[ "$1" =~ ^[1-9][0-9]?$ ]]; then
    echo "Error: value passed to --concurrency must be a number"
    exit 1
  fi
}

while [[ $# -gt 0 ]]
do
    key="$1"
    shift
    case ${key} in
        --concurrency|-c)
            validateConcurrency "$1"
            CONCURRENCY="$1"
            shift
            ;;
        -*)
            echo "Unknown flag ${key}"
            exit 1
            ;;
        *) # unknown option
            POSITIONAL+=("$key") # save it in an array for later
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters


source ${ROOT_PATH}/testing-common.sh

suiteName="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
echo "----------------------------"
echo "- Testing Kyma..."
echo "----------------------------"

kc="kubectl $(context_arg)"

${kc} get clustertestsuites.testing.kyma-project.io > /dev/null 2>&1
if [[ $? -eq 1 ]]
then
   echo "ERROR: script requires ClusterTestSuite CRD"
   exit 1
fi

matchTests="" # match all tests

${kc} get cm dex-config -n kyma-system -ojsonpath="{.data}" | grep --silent "#__STATIC_PASSWORDS__"
if [[ $? -eq 1 ]]
then
  # if static users are not available, do not execute tests which requires them
  matchTests=$(${kc} get testdefinitions --all-namespaces -l 'require-static-users!=true' -o=go-template='  selectors:
    matchNames:
{{- range .items}}
      - name: {{.metadata.name}}
        namespace: {{.metadata.namespace}}
{{- end}}')
  echo "WARNING: following tests will be skipped due to the lack of static users:"
  echo "$(${kc} get testdefinitions --all-namespaces -l 'require-static-users=true' -o=go-template --template='{{- range .items}}{{printf " - %s\n" .metadata.name}}{{- end}}')"
fi

# creates a config map which provides the testing bundles
injectTestingBundles
trap removeTestingBundles ERR EXIT

cat <<EOF | ${kc} apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: ${suiteName}
spec:
  maxRetries: 1
  concurrency: ${CONCURRENCY}
${matchTests}
EOF

startTime=$(date +%s)

testExitCode=0
previousPrintTime=-1

while true
do
    currTime=$(date +%s)
    statusSucceeded=$(${kc} get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$(${kc} get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$(${kc} get cts  ${suiteName} -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

    if [[ "${statusSucceeded}" == *"True"* ]]; then
       echo "Test suite '${suiteName}' succeeded."
       break
    fi

    if [[ "${statusFailed}" == *"True"* ]]; then
        echo "Test suite '${suiteName}' failed."
        testExitCode=1
        break
    fi

    if [[ "${statusError}" == *"True"* ]]; then
        echo "Test suite '${suiteName}' errored."
        testExitCode=1
        break
    fi

    sec=$((currTime-startTime))
    min=$((sec/60))
    if (( min > 60 )); then
        echo "Timeout for test suite '${suiteName}' occurred."
        testExitCode=1
        break
    fi
    if (( $previousPrintTime != $min )); then
        echo "ClusterTestSuite not finished. Waiting..."
        previousPrintTime=${min}
    fi
    sleep 3
done

echo "Test summary"
kubectl get cts  ${suiteName} -o=go-template --template='{{range .status.results}}{{printf "Test status: %s - %s" .name .status }}{{ if gt (len .executions) 1 }}{{ print " (Retried)" }}{{end}}{{print "\n"}}{{end}}'

waitForTerminationAndPrintLogs ${suiteName}
cleanupExitCode=$?

echo "ClusterTestSuite details:"
kubectl get cts ${suiteName} -oyaml

kubectl delete cts ${suiteName}

printImagesWithLatestTag
latestTagExitCode=$?

exit $(($testExitCode + $cleanupExitCode + $latestTagExitCode))