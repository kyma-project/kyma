#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# connectToMinIO connect to MinIO
#
# Arguments:
#   $1 - Minio endpoint
#   $2 - Minio port
#   $3 - Minio accessKey
#   $4 - Minio secretKey
connectToMinIO() {
  local -r minio_endpoint="${1}"
  local -r minio_port="${2}"
  local -r access_key="${3}"
  local -r secret_key="${4}"

  local attempts=0
  local -r local=29

  echo "Connecting to Minio server: http://${minio_endpoint}:${minio_port}" ;
  local -r mc_command="mc config host add myminio http://${minio_endpoint}:${minio_port} ${access_key} ${secret_key}" ;

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
#   $1 - Bucket name
checkBucketExists() {
  local -r bucket_name="${1}"
  CMD=$(mc ls myminio/${bucket_name} > /dev/null 2>&1)
  return $?
}

# createBucket create bucket with given name and policy type
#
# Arguments:
#   $1 - Bucket name
#   $2 - Type of policy. Available `none` (private) and `download` (public) values
createBucket() {
  local -r bucket_name="${1}"
  local -r policy="${2}"

  if ! checkBucketExists "${bucket_name}" ; then
    echo "Creating bucket '${bucket_name}'"
    mc mb "myminio/${bucket_name}"
  else
    echo "Bucket '${bucket_name}' already exists."
  fi

  echo "Setting policy of bucket '${bucket_name}' to '${policy}'."
  mc policy "${policy}" "myminio/${bucket_name}"
}

# copyToBucket copy buckets from temporary local storage to MinIO
#
# Arguments:
#   $1 - Bucket name
#   $2 - Type of policy. Available `none` (private) and `download` (public) values
copyToBucket() {
  local -r bucket_name="${1}"
  local -r policy="${2}"

  createBucket "${bucket_name}" "${policy}"
  echo "Copying to bucket '${bucket_name}'"
  mc mirror "${LOCAL_STORAGE}/${bucket_name}/" "myminio/${bucket_name}"
}

# copyToBucket copy buckets from MinIO to temporary local storage
#
# Arguments:
#   $1 - Bucket name
copyFromBucket() {
  local -r bucket_name="${1}"
  mkdir -p "${LOCAL_STORAGE}/${bucket_name}"

  if checkBucketExists "${bucket_name}"; then
    echo "Copying from bucket '${bucket_name}'"
    mc mirror "myminio/${bucket_name}" "${LOCAL_STORAGE}/${bucket_name}/"
  fi
}

main() {

}
main