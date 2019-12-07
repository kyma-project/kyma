#!/bin/sh

readonly ASSET_STORE_MINIO_HOST="assetstore"
readonly RAFTER_MINIO_HOST="rafter"

set -o errexit
set -o nounset
set -o pipefail

# connectToMinIO connect to MinIO by custom host
#
# Arguments:
#   $1 - MinIO host
#   $2 - Minio endpoint
#   $3 - Minio port
#   $4 - Minio accessKey
#   $5 - Minio secretKey
connectToMinIO() {
  local -r minio_host="${1}"
  local -r minio_endpoint="${2}"
  local -r minio_port="${3}"
  local -r access_key="${4}"
  local -r secret_key="${5}"

  local attempts=0
  local -r local=29

  echo "Connecting to Minio server: http://${minio_endpoint}:${minio_port}" ;
  local -r mc_command="mc config host add ${minio_host} http://${minio_endpoint}:${minio_port} ${access_key} ${secret_key}" ;

  $mc_command

  local status=$?
  until [ "${status}" = 0 ]
  do
    attempts=`expr ${attempts} + 1` ;
    echo "Failed attempts: ${attempts}"
    if [ "${attempts}" -gt "${local}" ]; then
      exit 1
    fi

    sleep 1
    $mc_command
    status=$?
  done

  return 0
}

# checkBucketExists check if bucket exists with given name
#
# Arguments:
#   $1 - MinIO host
#   $2 - Bucket name
checkBucketExists() {
  local -r minio_host="${1}"
  local -r bucket_name="${2}"

  CMD=$(mc ls ${minio_host}/${bucket_name} > /dev/null 2>&1)
  return $?
}

# createBucket create bucket with given name and policy type
#
# Arguments:
#   $1 - MinIO host
#   $2 - Bucket name
#   $3 - Type of policy. Available `none` (private) and `download` (public) values
createBucket() {
  local -r minio_host="${1}"
  local -r bucket_name="${2}"
  local -r policy="${3}"

  if ! checkBucketExists "${minio_host}" "${bucket_name}" ; then
    echo "Creating bucket '${bucket_name}'"
    mc mb "${minio_host}/${bucket_name}"
  else
    echo "Bucket '${bucket_name}' already exists."
  fi

  echo "Setting policy of bucket '${bucket_name}' to '${policy}'."
  mc policy "${policy}" "${minio_host}/${bucket_name}"
}

# copyToBucket copy bucket's content from temporary local storage to MinIO
#
# Arguments:
#   $1 - MinIO host
#   $2 - Bucket name
#   $3 - Type of policy. Available `none` (private) and `download` (public) values
copyToBucket() {
  local -r minio_host="${1}"
  local -r bucket_name="${2}"
  local -r policy="${3}"

  createBucket "${minio_host}" "${bucket_name}" "${policy}"
  echo "Copying to bucket '${bucket_name}'"
  mc mirror "${LOCAL_STORAGE}/${bucket_name}/" "${minio_host}/${bucket_name}"
}

# copyToBucket copy bucket's content from MinIO to temporary local storage
#
# Arguments:
#   $1 - MinIO host
#   $2 - Bucket name
copyFromBucket() {
  local -r minio_host="${1}"
  local -r bucket_name="${2}"

  mkdir -p "${LOCAL_STORAGE}/${bucket_name}"

  if checkBucketExists "${minio_host}" "${bucket_name}"; then
    echo "Copying from bucket '${bucket_name}'"
    mc mirror "${minio_host}/${bucket_name}" "${LOCAL_STORAGE}/${bucket_name}/"
  fi
}

# copyContentFromAssetStore copy AssetStore's MinIO content to temporary local storage
#
copyContentFromAssetStore() {
  connectToMinIO "${ASSET_STORE_MINIO_HOST}" "${ASSET_STORE_MINIO_ENDPOINT}" "${ASSET_STORE_MINIO_PORT}" "${ASSET_STORE_MINIO_ACCESS_KEY}" "${ASSET_STORE_MINIO_SECRET_KEY}"
  copyFromBucket "${ASSET_STORE_MINIO_HOST}" "${ASSET_STORE_PRIVATE_BUCKET}"
  copyFromBucket "${ASSET_STORE_MINIO_HOST}" "${ASSET_STORE_PUBLIC_BUCKET}"
}

# copyContentToRafter copy temporary local storage content to Rafter's MinIO  
#
copyContentToRafter() {
  connectToMinIO "${RAFTER_MINIO_HOST}" "${RAFTER_MINIO_ENDPOINT}" "${RAFTER_MINIO_PORT}" "${RAFTER_MINIO_ACCESS_KEY}" "${RAFTER_MINIO_SECRET_KEY}"
  copyToBucket "${RAFTER_MINIO_HOST}" "${ASSET_STORE_PRIVATE_BUCKET}" "none"
  copyToBucket "${RAFTER_MINIO_HOST}" "${ASSET_STORE_PUBLIC_BUCKET}" "download"
}

main() {
  copyContentFromAssetStore
  copyContentToRafter
}
main
