# Copies $SOURCE_SECRET to $TARGET_SECRET in $NAMESPACE. When copying finished successfully $SOURCE_SECRET is removed.
# The script is idempotent

set -e
set -o pipefail

echo "Copying ${SOURCE_SECRET} to ${TARGET_SECRET} started"

secret=$(kubectl -n "$NAMESPACE" get secret "${SOURCE_SECRET}" --ignore-not-found)
if [ -n "$secret" ]
then
    echo "Removing ${TARGET_SECRET}"
    kubectl delete secret ${TARGET_SECRET} -n=${NAMESPACE} --ignore-not-found

    echo "Fetching data"
    cacert=$(kubectl -n "${NAMESPACE}" get secret "${SOURCE_SECRET}" -o jsonpath='{.data.ca\.crt}' --ignore-not-found)
    cakey=$(kubectl -n "${NAMESPACE}" get secret "${SOURCE_SECRET}" -o jsonpath='{.data.ca\.key}' --ignore-not-found)

    echo "Creating ${TARGET_SECRET} secret"
    kubectl create secret generic "${TARGET_SECRET}" -n "${NAMESPACE}" --from-literal=ca.crt="$cacert" --from-literal=ca.key="$cakey"

    echo "Deleting ${SOURCE_SECRET} secret"
    kubectl delete secret ${SOURCE_SECRET} -n=${NAMESPACE}
 fi

echo "Copying ${SOURCE_SECRET} to ${TARGET_SECRET} done"