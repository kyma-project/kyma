set -eu

function rollOut() {
  DEPLOY=$(kubectl get deploy -n "${NAMESPACE}" -l "${1}" -o name)
  kubectl set env -n $NAMESPACE ${DEPLOY} sync=$(date "+%Y%m%d-%H%M%S")
  kubectl rollout status -n $NAMESPACE ${DEPLOY}
}

inotifywait -e DELETE_SELF -m "${WATCH_FILE}" |
   while read path _ file; do
       echo "---> $path$file modified"
       rollOut "${DEPLOYMENT_SELECTOR}"
   done
