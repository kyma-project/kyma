#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

CONCURRENCY=1
CLEANUP="true"

POSITIONAL=()

TEST_NAME=
TEST_NAMESPACE=

function validateConcurrency() {
  if [[ -z "$1" ]]; then
    echo "Error: --concurrency requires a value"
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
        --test-name|-t)
            TEST_NAME="$1"
            shift
            ;;
        --test-namespace|-tn)
            TEST_NAMESPACE="$1"
            shift
            ;;
        --cleanup)
            CLEANUP="$1"
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

existingCTSs=$(${kc} get cts -o=name)
for cts in ${existingCTSs}
do
  echo "Removing: ${cts}"
  ${kc} delete ${cts}
done

matchTests=$(${kc} get testdefinitions --all-namespaces -l 'kyma-project.io/upgrade-e2e-test!=executeTests' -o=go-template='  selectors:
    matchNames:
{{- range .items}}
      - name: {{.metadata.name}}
        namespace: {{.metadata.namespace}}
{{- end}}') # match all tests, ignore upgrade test

if [[ -n "${TEST_NAME}" && -n "${TEST_NAMESPACE}" ]]; then
  matchTests="  selectors:
    matchNames:
      - name: ${TEST_NAME}
        namespace: ${TEST_NAMESPACE}
"
fi

# creates a ClusterAddonsConfiguration which provides the testing addons
injectTestingAddons
if [[ $? -eq 1 ]]; then
  exit 1
fi

if [[ ${CLEANUP} = "true" ]]; then
  trap removeTestingAddons EXIT
fi

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

waitForTestSuiteResult ${suiteName}
testExitCode=$?

if [[ ${CLEANUP} = "true" ]]; then
  deleteCTS ${suiteName}
fi

printImagesWithLatestTag
latestTagExitCode=$?

exit $(($testExitCode + $latestTagExitCode))