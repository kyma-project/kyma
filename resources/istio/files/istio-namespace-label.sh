#!/bin/bash

# NAMESPACES - env variable, that contains space-separated
# list of namespaces that will be labelled by kubectl.
# If namespace does not exist it will be skipped.
# Other errors are not covered and will be populated if occur.
set -e
for namespace in $NAMESPACES ; do
    if RESULT=$(kubectl get namespace "$namespace" 2>&1) ; then
     echo "---> Setting label to $namespace"
     kubectl label namespace "$namespace" "istio-injection=disabled" --overwrite
    else
     NS_NOT_FOUND_ERROR="not found"
     if [[ $RESULT =~ .*"$NS_NOT_FOUND_ERROR".* ]] ; then
       echo "---> Namespace $namespace could not be labelled, as it was not found. Skipping."
     else
       echo "$RESULT"
       exit 1
     fi
    fi
done
