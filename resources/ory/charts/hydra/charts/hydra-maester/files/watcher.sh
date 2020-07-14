set -e

export SECRET_FILE=/etc/secrets/dsn
export NAMESPACE="{{ .Release.Namespace }}"

apk add inotify-tools

function rollOut() {
    local DEPLOYMENTS=( 
    $(kubectl get deploy -n $NAMESPACE -l "app.kubernetes.io/name=hydra" -o name)
  )

  for deploy in "${DEPLOYMENTS[@]}"; do
    kubectl set env -n $NAMESPACE ${deploy} sync=$(date "+%Y%m%d-%H%M%S")
      kubectl rollout status -n $NAMESPACE ${deploy}
  done
}

inotifywait -e DELETE_SELF -m $SECRET_FILE |
   while read path _ file; do
       echo "---> $path$file modified"
       rollOut
   done
