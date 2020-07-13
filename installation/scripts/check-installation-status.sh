#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

#files=$(find ${CURRENT_DIR}/../resources/helm-releases/ -type f -name '*.yaml' -maxdepth 1)
#for filename in $files; do
#  echo $(yq r $filename "spec.releaseName")
#done

function checkInstallationProgress() {
  totalNumberOfReleases=$(find ${CURRENT_DIR}/../resources/helm-releases/ -name '*.yaml' | wc -l)
  echo "Total number of releases to be installed : $totalNumberOfReleases"
  while :; do
    deployedReleases="$(($(helm ls --all-namespaces --deployed | tail -n +2 | wc -l) - 1))"
    if [ $totalNumberOfReleases != $deployedReleases ]; then
      echo "Still not ready, number of already deployed releases : $deployedReleases"
      echo "waiting for : "
      echo "$(helm ls --all-namespaces --pending --failed)"
      sleep 1s
    else
      echo "All required releases are installed"
      break
    fi
  done

}
checkInstallationProgress
