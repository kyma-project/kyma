#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${ROOT_PATH}/testing-common.sh"

if [ -z "${DOMAIN}" ] ; then
  echo "ERROR: DOMAIN is not set"
  exit 1
fi

cleanupHelmE2ERelease () {
    local release=$1
    log 'Running cleanup'
    helm del --purge "${release}" --tls
    local deleteErr=$?
    if [ ${deleteErr} -ne 0 ]
    then
      log "FAILED cleaning release. This might be because the release hasn't been installed yet \n" red
      return 1
    fi
    while helm list --deleting --tls 2>/dev/null | grep "${release}" ; do
        sleep 1
        echo .
    done
}

# creates a config map which provides the testing bundles	
injectTestingBundles	

testcase="${ROOT_PATH}"/../../tests/end-to-end/backup-restore-test/deploy/chart/backup-test
release=$(basename "$testcase")

ADMIN_EMAIL=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode)
ADMIN_PASSWORD=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode)

helm install "$testcase" --name "${release}" --namespace end-to-end --set global.ingress.domainName="${DOMAIN}" --set-file global.adminEmail=<(echo -n "${ADMIN_EMAIL}") --set-file global.adminPassword=<(echo -n "${ADMIN_PASSWORD}") --tls

suiteName="testsuite-backup-$(date '+%Y-%m-%d-%H-%M')"
echo "----------------------------"
echo "- Testing Kyma Backup and Restore functionality..."
echo "----------------------------"

kc="kubectl $(context_arg)"

${kc} get cts > /dev/null 2>&1
if [[ $? -eq 1 ]]
then
   echo "ERROR: script requires ClusterTestSuite CRD"
   exit 1
fi

matchTests=$(${kc} get testdefinitions --all-namespaces -l 'app=backup-test' -o=go-template='  selectors:
    matchNames:
{{- range .items}}
      - name: {{.metadata.name}}
        namespace: {{.metadata.namespace}}
{{- end}}')

cat <<EOF | ${kc} apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: ${suiteName}
spec:
  maxRetries: 1
  concurrency: 1
${matchTests}
EOF

startTime=$(date +%s)

testExitCode=0
previousPrintTime=-1

while true
do
    currTime=$(date +%s)
    statusSucceeded=$(${kc} get cts "${suiteName}"  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$(${kc} get cts "${suiteName}"  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$(${kc} get cts  "${suiteName}" -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

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
    if (( previousPrintTime != min )); then
        echo "ClusterTestSuite not finished. Waiting..."
        previousPrintTime=${min}
    fi
    sleep 3
done

echo "Test summary"
kubectl get cts  "${suiteName}" -o=go-template --template='{{range .status.results}}{{printf "Test status: %s - %s" .name .status }}{{ if gt (len .executions) 1 }}{{ print " (Retried)" }}{{end}}{{print "\n"}}{{end}}'

waitForTerminationAndPrintLogs "${suiteName}"
cleanupExitCode=$?

echo "ClusterTestSuite details:"
kubectl get cts "${suiteName}" -oyaml

kubectl delete cts "${suiteName}"

cleanupHelmE2ERelease "${release}"
releaseCleanupResult=$?

exit $((testExitCode + cleanupExitCode + releaseCleanupResult))
