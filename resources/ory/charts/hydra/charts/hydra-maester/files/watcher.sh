set -e

export SECRET_FILE=/etc/secrets/dsn
export NAMESPACE="{{ .Release.Namespace }}"

function rollOut() {
  DEPLOY=$(kubectl get deploy -n $NAMESPACE -l "app.kubernetes.io/name=${1}" -o name)
  kubectl set env -n $NAMESPACE ${DEPLOY} sync=$(date "+%Y%m%d-%H%M%S")
  kubectl rollout status -n $NAMESPACE ${DEPLOY}
}

inotifywait -e DELETE_SELF -m $SECRET_FILE |
   while read path _ file; do
       echo "---> $path$file modified"
       rollOut "hydra"
       rollOut "hydra-maester"
   done
