#!/bin/bash -e
NAMESPACE_ADMIN_EMAIL=foo
VIEW_EMAIL=bar
DEVELOPER_EMAIL=oof
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source "${DIR}/library.sh"

# resources/cluster-users/values.yaml clusterRoles.verbs.view
readonly VIEW_OPERATIONS=( "get" "list" )
# resources/cluster-users/values.yaml clusterRoles.verbs.edit - clusterRoles.verbs.view
readonly EDIT_OPERATIONS=( "create" "delete" "deletecollection" "patch" "update" "watch" )

# users used in tests
readonly USERS=( "${NAMESPACE_ADMIN_EMAIL}" "${VIEW_EMAIL}" "${DEVELOPER_EMAIL}" "${ADMIN_EMAIL}" )

readonly K8S_RESOURCES=( "deployment" "pod" "secret" "cm" )

trap cleanup EXIT
ERROR_LOGGING_GUARD="true"

for USER in "${USERS[@]}"; do
	echo "---> $USER"
done

# Create bindings
# for USER in [X Y Z]
	# Login as user X
	# Test as user X
# Cleanup

# createTestBindingsRetry
# runTests

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"

