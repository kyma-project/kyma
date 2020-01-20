#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${ROOT_PATH}/testing-common.sh"

if [ -z "${DOMAIN}" ] ; then
  echo "ERROR: DOMAIN is not set"
  exit 1
fi

# ACTION must be one of the following: testBeforeBackup, testAfterRestore
if [ -z "${ACTION}" ] ; then
  echo "ERROR: ACTION is not set"
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
if [[ "${ACTION}" == "testBeforeBackup" ]]; then
  injectTestingAddons
  job=before-backup
else
  job=after-restore
fi  

testcase="${ROOT_PATH}"/../../tests/end-to-end/backup/chart/backup-test
release=$(basename "$testcase")

ADMIN_EMAIL=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode)
ADMIN_PASSWORD=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode)


echo helm install "$testcase" --name "${release}" --namespace backup-test --set global.ingress.domainName="${DOMAIN}" --set job="${job}" --set-file global.adminEmail=<(echo -n "${ADMIN_EMAIL}") --set-file global.adminPassword=<(echo -n "${ADMIN_PASSWORD}") --tls
helm install "$testcase" --name "${release}" --namespace backup-test --set global.ingress.domainName="${DOMAIN}" --set job="${job}" --set-file global.adminEmail=<(echo -n "${ADMIN_EMAIL}") --set-file global.adminPassword=<(echo -n "${ADMIN_PASSWORD}") --tls

suiteName="testsuite-backup-$(date '+%Y-%m-%d-%H-%M')"
echo "---------------------------------------------------"
echo "- Running backup restore tests with action set to ${ACTION}..."
echo "---------------------------------------------------"

kc="kubectl $(context_arg)"

${kc} get cts > /dev/null 2>&1
if [[ $? -eq 1 ]]
then
   echo "ERROR: script requires ClusterTestSuite CRD"
   exit 1
fi

matchTests=$(${kc} get testdefinitions --all-namespaces -l 'app=e2e-backup-test' -o=go-template='  selectors:
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
  maxRetries: 0
  concurrency: 1
${matchTests}
EOF

waitForTestSuiteResult ${suiteName}
testExitCode=$?

kubectl delete cts "${suiteName}"

cleanupHelmE2ERelease "${release}"
releaseCleanupResult=$?

exit $((testExitCode + releaseCleanupResult))
