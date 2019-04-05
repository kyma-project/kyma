#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${ROOT_PATH}/utils.sh"
source "${ROOT_PATH}/testing-common.sh"

if [ -z "${DOMAIN}" ] ; then
  echo "ERROR: DOMAIN is not set"
  exit 1
fi

cleanupHelmE2ERelease () {
    local release=$1
    log 'Running cleanup'

    set +o errexit
    if [  -f "$(helm home)/ca.pem" ]; then
        local HELM_ARGS="--tls"
    fi
    helm del --purge "${release}" ${HELM_ARGS}
    local deleteErr=$?
    if [ ${deleteErr} -ne 0 ]
    then
      log "FAILED cleaning release.\n" red
      return 1
    fi
    while helm list --deleting ${HELM_ARGS} 2>/dev/null | grep "${release}" ; do
        sleep 1
        echo .
    done
}


echo "-------------------------------"
echo "- Ensure test Pods are deleted "
echo "-------------------------------"

cleanupHelmTestPods end-to-end
cleanupE2EErr=$?

if [ ${cleanupE2EErr} -ne 0 ]
then
    exit 1
fi

echo "----------------------------"
echo "- E2E Testing Kyma..."
echo "----------------------------"

exitCode=0


testcase="${ROOT_PATH}"/../../tests/end-to-end/backup-restore-test/deploy/chart/backup-test
release=$(basename "$testcase")

cleanupHelmE2ERelease "${release}"

ADMIN_EMAIL=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode)
ADMIN_PASSWORD=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode)

set +o errexit
if [  -f "$(helm home)/ca.pem" ]; then
    local HELM_ARGS="--tls"
fi

helm install "$testcase" --name "${release}" --namespace end-to-end --set global.ingress.domainName="${DOMAIN}" --set-file global.adminEmail=<(echo -n "${ADMIN_EMAIL}") --set-file global.adminPassword=<(echo -n "${ADMIN_PASSWORD}") ${HELM_ARGS}
helm test "${release}" --timeout 10000 ${HELM_ARGS}
testResult=$?
if [ $testResult -eq 0 ]
then
    releasesToClean="$releasesToClean ${release}"
else
    exitCode=$testResult
fi

checkAndCleanupTest end-to-end
cleanupResult=$?
if [ $cleanupResult -ne 0 ]
then
   exitCode=$cleanupResult
fi


for release in $releasesToClean; do
    cleanupHelmE2ERelease "${release}"
    cleanupResult=$?
    if [ $cleanupResult -ne 0 ]
    then
        exitCode=$cleanupResult
    fi
done

if [ ${exitCode} -ne 0 ]
then
    log FAIL red
    exit 1
else
    log SUCCESS green
    exit 0
fi
